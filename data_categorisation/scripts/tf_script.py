#!/usr/bin/env python3
import json
import re
import sys
from typing import Union

ALLOWED_R_TYPES = {
    "aws_s3_bucket",
    "aws_dynamodb_table",
    "aws_db_instance",
    "aws_rds_cluster",
}

TAG_KEY_CLASS = "data_classification"
TAG_KEY_CAT   = "data_category"
TAG_VAL_CLASS = "{{resolve:ssm:/DataClassification/CHANGE_ME}}"
TAG_VAL_CAT   = "{{resolve:ssm:/DataCategory/CHANGE_ME}}"

# ------------------ newline helpers ------------------
def _pick_write_newline(seen: Union[None, str, tuple]) -> str:
    if seen is None:
        return "\n"
    if isinstance(seen, tuple):
        return "\r\n" if "\r\n" in seen else (seen[0] if isinstance(seen[0], str) else "\n")
    return seen

def _had_final_newline(raw_text: str) -> bool:
    return raw_text.endswith("\n") or raw_text.endswith("\r")

# ------------------ .tf (HCL) patching ------------------
RES_START_RE = re.compile(
    r'^(?P<indent>[ \t]*)resource\s+"(?P<rtype>aws_s3_bucket|aws_dynamodb_table|aws_db_instance|aws_rds_cluster)"\s+"(?P<name>[^"]+)"\s*\{\s*$'
)
TAGS_OPEN_RE = re.compile(r'^(?P<indent>[ \t]*)tags\s*=\s*\{\s*$')
MAP_OPEN = "{"
MAP_CLOSE = "}"

def _block_end(lines, start_idx):
    depth = 0
    for i in range(start_idx, len(lines)):
        line = lines[i]
        depth += line.count(MAP_OPEN)
        depth -= line.count(MAP_CLOSE)
        if depth == 0:
            return i
    return None

def _tags_body_bounds(lines, start_idx):
    m = TAGS_OPEN_RE.match(lines[start_idx])
    if not m:
        return None, None, None
    indent = m.group("indent")
    end = _block_end(lines, start_idx)
    if end is None:
        return None, None, None
    body_start = start_idx + 1
    body_end = end - 1  # last line before the closing brace
    return indent, body_start, body_end

def _has_key_in_tags(lines, body_start, body_end, key):
    key_pat = re.compile(rf'^\s*("?{re.escape(key)}"?)\s*=')
    for i in range(body_start, body_end + 1):
        if key_pat.search(lines[i]):
            return True
    return False

def _insert_tag_pairs(lines, insert_at, entry_indent):
    ins = [
        f'{entry_indent}{TAG_KEY_CLASS} = "{TAG_VAL_CLASS}"',
        f'{entry_indent}{TAG_KEY_CAT} = "{TAG_VAL_CAT}"',
    ]
    # insert as a block at the top of the tags map
    return lines[:insert_at] + ins + lines[insert_at:]

def patch_tf(path: str) -> bool:
    try:
        with open(path, "r", encoding="utf-8", newline=None) as fh:
            raw = fh.read()
            nl_seen = fh.newlines
        lines = raw.splitlines()
        had_nl = _had_final_newline(raw)
    except Exception:
        print(f"Skip (unreadable .tf): {path}")
        return False

    out = []
    i = 0
    changed = False

    while i < len(lines):
        m = RES_START_RE.match(lines[i])
        if not m:
            out.append(lines[i]); i += 1; continue

        indent = m.group("indent")
        r_end = _block_end(lines, i)
        if r_end is None:
            # malformed block, copy-through
            out.append(lines[i]); i += 1; continue

        body = lines[i+1:r_end]  # inside the resource braces
        # find an existing tags = { ... } in the body
        j = i + 1
        found_tags_line = None
        while j < r_end:
            if TAGS_OPEN_RE.match(lines[j]):
                found_tags_line = j
                break
            j += 1

        if found_tags_line is None:
            # create new tags block right after the opening line
            child_indent = indent + "  "
            entry_indent = child_indent + "  "
            new_bits = [
                f"{child_indent}tags = {{",
                f'{entry_indent}{TAG_KEY_CLASS} = "{TAG_VAL_CLASS}"',
                f'{entry_indent}{TAG_KEY_CAT} = "{TAG_VAL_CAT}"',
                f"{child_indent}}}",
            ]
            out.append(lines[i])
            out.extend(new_bits)
            out.extend(body)
            out.append(lines[r_end])
            changed = True
            i = r_end + 1
            continue

        # we have a tags map, insert missing keys at its top
        tag_indent, b_start, b_end = _tags_body_bounds(lines, found_tags_line)
        if tag_indent is None:
            # broken tags map, copy-through
            out.extend(lines[i:r_end+1])
            i = r_end + 1
            continue

        have_class = _has_key_in_tags(lines, b_start, b_end, TAG_KEY_CLASS)
        have_cat   = _has_key_in_tags(lines, b_start, b_end, TAG_KEY_CAT)

        if have_class and have_cat:
            out.extend(lines[i:r_end+1])
            i = r_end + 1
            continue

        entry_indent = tag_indent + "  "
        # insert immediately after opening brace line
        insert_at = found_tags_line + 1
        new_chunk = lines[i:r_end+1]
        ins_lines = []
        if not have_class:
            ins_lines.append(f'{entry_indent}{TAG_KEY_CLASS} = "{TAG_VAL_CLASS}"')
        if not have_cat:
            ins_lines.append(f'{entry_indent}{TAG_KEY_CAT} = "{TAG_VAL_CAT}"')
        new_chunk = new_chunk[:insert_at - i] + ins_lines + new_chunk[insert_at - i:]
        out.extend(new_chunk)
        changed = True
        i = r_end + 1

    if changed:
        write_nl = _pick_write_newline(nl_seen)
        tail = "\n" if had_nl else ""
        with open(path, "w", encoding="utf-8", newline=write_nl) as fh:
            fh.write("\n".join(out) + tail)
        print(f"Patched (TF): {path}")
    else:
        print(f"No matching resources or tags already present (.tf): {path}")
    return changed

# ------------------ .tf.json patching ------------------
def patch_tfjson(path: str) -> bool:
    try:
        with open(path, "r", encoding="utf-8", newline=None) as fh:
            raw = fh.read()
            nl_seen = fh.newlines
        had_nl = _had_final_newline(raw)
        data = json.loads(raw)
    except Exception:
        print(f"Skip (unreadable .tf.json): {path}")
        return False

    changed = False
    res = data.get("resource", {})
    if isinstance(res, dict):
        for rtype in ALLOWED_R_TYPES:
            if rtype not in res:
                continue
            r_objs = res.get(rtype)
            if not isinstance(r_objs, dict):
                continue
            for name, body in r_objs.items():
                if not isinstance(body, dict):
                    continue
                tags = body.get("tags")
                if tags is None:
                    body["tags"] = {
                        TAG_KEY_CLASS: TAG_VAL_CLASS,
                        TAG_KEY_CAT: TAG_VAL_CAT,
                    }
                    changed = True
                elif isinstance(tags, dict):
                    add = False
                    if TAG_KEY_CLASS not in tags:
                        tags[TAG_KEY_CLASS] = TAG_VAL_CLASS
                        add = True
                    if TAG_KEY_CAT not in tags:
                        tags[TAG_KEY_CAT] = TAG_VAL_CAT
                        add = True
                    changed = changed or add
                else:
                    # tags is not a map; leave it unchanged
                    pass

    if changed:
        write_nl = _pick_write_newline(nl_seen)
        tail = "\n" if had_nl else ""
        with open(path, "w", encoding="utf-8", newline=write_nl) as fh:
            json.dump(data, fh, indent=2, ensure_ascii=False)
            fh.write(tail)
        print(f"Patched (.tf.json): {path}")
    else:
        print(f"No matching resources or tags already present (.tf.json): {path}")
    return changed

# ------------------ CLI ------------------
def main(paths):
    any_changed = False
    for p in paths:
        lp = p.lower()
        if lp.endswith(".tf"):
            any_changed |= patch_tf(p)
        elif lp.endswith(".tf.json"):
            any_changed |= patch_tfjson(p)
        else:
            print(f"Skip (unknown type): {p}")
    return 0

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("usage: tf_tags.py <file1> [file2 ...]", file=sys.stderr)
        sys.exit(2)
    sys.exit(main(sys.argv[1:]))
