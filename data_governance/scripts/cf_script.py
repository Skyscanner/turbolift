#!/usr/bin/env python3
import re
import sys
from typing import Union

ALLOWED_TYPES = {
    "AWS::S3::Bucket",
    "AWS::DynamoDB::Table",
    "AWS::DynamoDB::GlobalTable",
    "AWS::RDS::DBInstance",
    "AWS::RDS::DBCluster",
}

TAG_KEY_CLASS = "data_classification"
TAG_KEY_CAT   = "data_category"
TAG_VAL_CLASS = "CHANGE_ME"
TAG_VAL_CAT   = "CHANGE_ME"

def _indent_len(s: str) -> int:
    return len(s) - len(s.lstrip(" "))

def _pick_write_newline(seen: Union[None, str, tuple]) -> str:
    if seen is None:
        return "\n"
    if isinstance(seen, tuple):
        return "\r\n" if "\r\n" in seen else (seen[0] if isinstance(seen[0], str) else "\n")
    return seen

def _had_final_newline(raw_text: str) -> bool:
    return raw_text.endswith("\n") or raw_text.endswith("\r")

YAML_PROPS_LINE_RE = re.compile(r'^(?P<indent>[ \t]*)Properties\s*:\s*$')
YAML_TYPE_LINE_RE  = re.compile(r'^(?P<indent>[ \t]*)Type\s*:\s*(?P<rtype>.+?)\s*$')
YAML_TAGS_LINE_RE  = re.compile(r'^(?P<indent>[ \t]*)Tags\s*:\s*$')
YAML_REPLICAS_LINE_RE = re.compile(r'^(?P<indent>[ \t]*)Replicas\s*:\s*$')

YAML_ITEM_KEY_RE   = re.compile(r'^(?P<indent>[ \t]*)-\s*Key\s*:\s*')
YAML_LIST_ITEM_RE  = re.compile(r'^(?P<indent>[ \t]*)-\s+')

def _unquote_yaml_scalar(s: str) -> str:
    s = s.strip()
    if len(s) >= 2 and ((s[0] == s[-1] == "'") or (s[0] == s[-1] == '"')):
        return s[1:-1]
    return s

def _yaml_block_end(lines, start_idx, parent_indent, max_look=5000):
    end = start_idx + 1
    limit = min(len(lines), start_idx + max_look)
    parent_len = len(parent_indent)
    while end < limit:
        l = lines[end]
        if l.strip() == "" or l.lstrip().startswith("#"):
            end += 1
            continue
        if _indent_len(l) <= parent_len:
            break
        end += 1
    return end

def _yaml_find_type_around(lines, i_props, props_indent):
    # search upward
    j = i_props - 1
    while j >= 0:
        m = YAML_TYPE_LINE_RE.match(lines[j])
        if m and m.group("indent") == props_indent:
            return _unquote_yaml_scalar(m.group("rtype"))
        if _indent_len(lines[j]) < len(props_indent):
            break
        j -= 1
    # search downward
    k = i_props + 1
    while k < len(lines):
        if lines[k].strip() and _indent_len(lines[k]) < len(props_indent):
            break
        m = YAML_TYPE_LINE_RE.match(lines[k])
        if m and m.group("indent") == props_indent:
            return _unquote_yaml_scalar(m.group("rtype"))
        k += 1
    return None

def _yaml_detect_item_indent(window, tags_idx, tags_indent):
    look = window[tags_idx+1 : min(len(window), tags_idx+25)]
    for l in look:
        if not l.strip():
            continue
        if l.startswith(tags_indent + "  -"):
            return tags_indent + "  "
        if l.startswith(tags_indent + "-"):
            return tags_indent
        break
    return tags_indent + "  "

def _yaml_collect_tags_block(window, tags_idx, item_indent):
    sub_start = tags_idx + 1
    sub_end = sub_start
    while sub_end < len(window):
        l = window[sub_end]
        if l.strip() == "":
            break
        if not l.startswith(item_indent):
            break
        after = l[len(item_indent):]
        if after and not after.startswith("-") and not after.startswith(" "):
            break
        sub_end += 1
    return sub_start, sub_end

def _yaml_pick_list_indent(existing, item_indent):
    for w in existing:
        mm = YAML_ITEM_KEY_RE.match(w)
        if mm:
            return mm.group("indent")
    for w in existing:
        mm = YAML_LIST_ITEM_RE.match(w)
        if mm:
            return mm.group("indent")
    return item_indent

def _yaml_has_key(line: str, key: str) -> bool:
    # Matches: - Key: data_classification     with optional quotes and extra spaces
    pat = rf'^\s*-\s*Key\s*:\s*(?:"{re.escape(key)}"|\'{re.escape(key)}\'|{re.escape(key)})\s*$'
    return re.match(pat, line) is not None

def _yaml_patch_or_create_tags_at_indent(window, desired_tags_indent):
    """
    Ensure a Tags: list exists at desired_tags_indent inside the given window, and inject our two tag pairs if missing.
    Returns (new_window, changed).
    """
    # Locate an existing `Tags:` at the desired indent
    tags_idx = None
    for j, w in enumerate(window):
        mt = YAML_TAGS_LINE_RE.match(w)
        if mt and mt.group("indent") == desired_tags_indent:
            tags_idx = j
            break

    def make_inserts(list_indent: str, val_indent: str):
        return [
            f"{list_indent}- Key: {TAG_KEY_CLASS}",
            f'{val_indent}Value: "{TAG_VAL_CLASS}"',
            f"{list_indent}- Key: {TAG_KEY_CAT}",
            f'{val_indent}Value: "{TAG_VAL_CAT}"',
        ]

    if tags_idx is None:
        tags_indent = desired_tags_indent
        list_indent = tags_indent + "  "
        val_indent  = list_indent + "  "
        new_block = [f"{tags_indent}Tags:"] + make_inserts(list_indent, val_indent)
        return new_block + window, True

    tags_indent = desired_tags_indent
    item_indent = _yaml_detect_item_indent(window, tags_idx, tags_indent)
    sub_start, sub_end = _yaml_collect_tags_block(window, tags_idx, item_indent)
    existing = window[sub_start:sub_end]

    has_class = any(_yaml_has_key(w, TAG_KEY_CLASS) for w in existing)
    has_cat   = any(_yaml_has_key(w, TAG_KEY_CAT) for w in existing)
    if has_class and has_cat:
        return window, False

    list_indent = _yaml_pick_list_indent(existing, item_indent)
    val_indent  = list_indent + "  "

    to_insert = []
    if not has_class:
        to_insert += [
            f"{list_indent}- Key: {TAG_KEY_CLASS}",
            f'{val_indent}Value: "{TAG_VAL_CLASS}"',
        ]
    if not has_cat:
        to_insert += [
            f"{list_indent}- Key: {TAG_KEY_CAT}",
            f'{val_indent}Value: "{TAG_VAL_CAT}"',
        ]

    # Insert at the top of the Tags list
    new_window = window[:sub_start] + to_insert + existing + window[sub_end:]
    return new_window, True

def _yaml_patch_or_create_tags(window, props_indent):
    # Properties-level Tags:
    desired_tags_indent = props_indent + "  "
    return _yaml_patch_or_create_tags_at_indent(window, desired_tags_indent)

def _yaml_collect_list_item_block(lines, start_idx, item_indent):
    """
    Collect a single list item block beginning at start_idx where lines[start_idx] starts with item_indent + '-'.
    Returns the end index (exclusive).
    """
    end = start_idx + 1
    n = len(lines)
    while end < n:
        l = lines[end]
        if l.strip() == "" or l.lstrip().startswith("#"):
            end += 1
            continue
        # next sibling item
        if l.startswith(item_indent + "-"):
            break
        # dedent out of the list
        if _indent_len(l) < len(item_indent):
            break
        end += 1
    return end

def _yaml_patch_replicas_tags(window, props_indent):
    """
    Find `Replicas:` under Properties, then ensure each replica item has Tags list with our keys.
    """
    child_indent = props_indent + "  "
    replicas_idx = None
    for j, w in enumerate(window):
        mr = YAML_REPLICAS_LINE_RE.match(w)
        if mr and mr.group("indent") == child_indent:
            replicas_idx = j
            break
    if replicas_idx is None:
        return window, False

    # Determine list item indent under `Replicas:`
    item_indent = child_indent + "  "
    k = replicas_idx + 1
    while k < len(window):
        l = window[k]
        if not l.strip() or l.lstrip().startswith("#"):
            k += 1
            continue
        if l.startswith(child_indent + "  -"):
            item_indent = child_indent + "  "
        elif l.startswith(child_indent + "-"):
            item_indent = child_indent
        break

    # Iterate each replica list item
    i = replicas_idx + 1
    changed = False
    out = window[:replicas_idx+1]
    while i < len(window):
        l = window[i]
        if not l.strip() or l.lstrip().startswith("#"):
            out.append(l); i += 1; continue
        if not l.startswith(item_indent + "-"):
            # left the Replicas list
            out.extend(window[i:])
            break

        li_start = i
        li_end = _yaml_collect_list_item_block(window, li_start, item_indent)
        replica_block = window[li_start:li_end]

        # A replica item has its child mapping at one indent deeper than item_indent
        replica_child_indent = item_indent + "  "
        patched_block, did = _yaml_patch_or_create_tags_at_indent(replica_block, replica_child_indent)
        if did:
            changed = True
        out.extend(patched_block)
        i = li_end

    return out, changed

def patch_yaml(path: str) -> bool:
    try:
        with open(path, "r", encoding="utf-8", newline=None) as fh:
            raw = fh.read()
            nl_seen = fh.newlines
        lines = raw.splitlines()
        had_nl = _had_final_newline(raw)
    except Exception:
        print(f"Skip (unreadable YAML): {path}")
        return False

    out, i, changed = [], 0, False
    while i < len(lines):
        m_props = YAML_PROPS_LINE_RE.match(lines[i])
        if not m_props:
            out.append(lines[i]); i += 1; continue

        props_indent = m_props.group("indent")
        end = _yaml_block_end(lines, i, props_indent)
        window = lines[i+1:end]
        rtype = _yaml_find_type_around(lines, i, props_indent)

        if rtype not in ALLOWED_TYPES:
            out.append(lines[i]); out.extend(window); i = end; continue

        # For GlobalTable, tags live under each replica item, not at Properties root
        if rtype == "AWS::DynamoDB::GlobalTable":
            new_window, did_change = _yaml_patch_replicas_tags(window, props_indent)
        else:
            new_window, did_change = _yaml_patch_or_create_tags(window, props_indent)

        out.append(lines[i]); out.extend(new_window)
        changed |= did_change
        i = end

    if changed:
        write_nl = _pick_write_newline(nl_seen)
        # join with '\n' and let TextIOWrapper translate to write_nl
        tail = "\n" if had_nl else ""
        with open(path, "w", encoding="utf-8", newline=write_nl) as fh:
            fh.write("\n".join(out) + tail)
        print(f"Patched (YAML): {path}")
    else:
        print(f"No data storage definitions found (YAML): {path}")
    return changed

JSON_PROPS_OPEN_RE = re.compile(r'^(?P<indent>[ \t]*)"Properties"\s*:\s*\{\s*$')
JSON_TYPE_LINE_RE  = re.compile(r'^(?P<indent>[ \t]*)"Type"\s*:\s*"(?P<rtype>[^"]+)"\s*,?\s*$')
JSON_TAGS_OPEN_RE  = re.compile(r'^(?P<indent>[ \t]*)"Tags"\s*:\s*\[\s*$')

JSON_REPLICAS_OPEN_RE = re.compile(r'^(?P<indent>[ \t]*)"Replicas"\s*:\s*\[\s*$')
JSON_OBJ_OPEN_RE   = re.compile(r'^(?P<indent>[ \t]*)\{')

def _json_obj_end(lines, start_idx):
    end, depth = start_idx + 1, 1
    while end < len(lines):
        depth += lines[end].count("{")
        depth -= lines[end].count("}")
        if depth == 0:
            return end
        end += 1
    return None

def _json_array_end(lines, start_idx):
    end, depth = start_idx + 1, 1
    while end < len(lines):
        depth += lines[end].count("[")
        depth -= lines[end].count("]")
        if depth == 0:
            return end
        end += 1
    return None

def _json_nearest_type(lines, i_props, props_indent, props_end):
    # search upward
    j = i_props - 1
    while j >= 0:
        m = JSON_TYPE_LINE_RE.match(lines[j])
        if m and m.group("indent") == props_indent:
            return m.group("rtype")
        if _indent_len(lines[j]) < len(props_indent):
            break
        j -= 1
    # search downward
    k = props_end + 1
    while k < len(lines):
        if lines[k].strip() and _indent_len(lines[k]) < len(props_indent):
            break
        m = JSON_TYPE_LINE_RE.match(lines[k])
        if m and m.group("indent") == props_indent:
            return m.group("rtype")
        k += 1
    return None

def _json_pick_obj_indent(arr_body, child_indent):
    for w in arr_body:
        mm = JSON_OBJ_OPEN_RE.match(w)
        if mm:
            return mm.group("indent")
    return child_indent + "  "

def _json_has_key(line: str, key: str) -> bool:
    # Matches: "Key": "data_classification"
    pat = rf'"Key"\s*:\s*"(?:{re.escape(key)})"'
    return re.search(pat, line) is not None

def _json_patch_or_create_tags(window, child_indent):
    tags_open = None
    for j, w in enumerate(window):
        mt = JSON_TAGS_OPEN_RE.match(w)
        if mt and mt.group("indent") == child_indent:
            tags_open = j
            break

    def make_block(child_indent_: str, obj_indent_: str, trailing_comma: bool) -> list[str]:
        block = [
            f'{child_indent_}"Tags": [',
            f"{obj_indent_}{{",
            f'{obj_indent_}  "Key": "{TAG_KEY_CLASS}",',
            f'{obj_indent_}  "Value": "{TAG_VAL_CLASS}"',
            f"{obj_indent_}}},",
            f"{obj_indent_}{{",
            f'{obj_indent_}  "Key": "{TAG_KEY_CAT}",',
            f'{obj_indent_}  "Value": "{TAG_VAL_CAT}"',
            f"{obj_indent_}}}",
            f"{child_indent_}]"+ ("," if trailing_comma else "")
        ]
        return block

    if tags_open is None:
        obj_indent = child_indent + "  "
        trailing = any(w.strip() for w in window)  # comma after ] if more properties follow
        block = make_block(child_indent, obj_indent, trailing)
        return block + window, True

    arr_end = _json_array_end(window, tags_open)
    if arr_end is None:
        return window, False

    body = window[tags_open+1:arr_end]
    has_dc  = any(_json_has_key(w, TAG_KEY_CLASS) for w in body)
    has_cat = any(_json_has_key(w, TAG_KEY_CAT) for w in body)
    if has_dc and has_cat:
        return window, False

    obj_indent = _json_pick_obj_indent(body, child_indent)

    inserts = []
    if not has_dc:
        inserts.append([
            f"{obj_indent}{{",
            f'{obj_indent}  "Key": "{TAG_KEY_CLASS}",',
            f'{obj_indent}  "Value": "{TAG_VAL_CLASS}"',
            f"{obj_indent}}}",
        ])
    if not has_cat:
        inserts.append([
            f"{obj_indent}{{",
            f'{obj_indent}  "Key": "{TAG_KEY_CAT}",',
            f'{obj_indent}  "Value": "{TAG_VAL_CAT}"',
            f"{obj_indent}}}",
        ])

    nonempty = any(w.strip() for w in body)
    new_body = []
    for idx, blk in enumerate(inserts):
        b = blk[:]
        if idx < len(inserts) - 1:
            b[-1] = b[-1] + ","
        new_body.extend(b)
    if nonempty:
        new_body[-1] = new_body[-1] + ","
    new_body.extend(body)
    return window[:tags_open+1] + new_body + window[arr_end:], True

def _json_patch_or_create_tags_in_obj(obj_body, obj_child_indent):
    """
    Ensure "Tags": [ ... ] exists inside a JSON object body with keys at obj_child_indent.
    obj_body excludes the surrounding { } lines.
    """
    return _json_patch_or_create_tags(obj_body, obj_child_indent)

def _json_patch_replicas_tags(window, child_indent):
    """
    Under Properties window, find "Replicas": [ ... ], then ensure each object inside has Tags.
    """
    rep_open = None
    for j, w in enumerate(window):
        mr = JSON_REPLICAS_OPEN_RE.match(w)
        if mr and mr.group("indent") == child_indent:
            rep_open = j
            break
    if rep_open is None:
        return window, False

    rep_end = _json_array_end(window, rep_open)
    if rep_end is None:
        return window, False

    body = window[rep_open+1:rep_end]
    out = []
    i = 0
    changed = False
    while i < len(body):
        l = body[i]
        if not l.strip():
            out.append(l); i += 1; continue
        mm = JSON_OBJ_OPEN_RE.match(l)
        if not mm:
            out.append(l); i += 1; continue
        # collect object
        obj_start = i
        depth, j = 1, i + 1
        while j < len(body) and depth > 0:
            depth += body[j].count("{")
            depth -= body[j].count("}")
            j += 1
        obj_end = j  # index after closing brace
        obj = body[obj_start:obj_end]

        open_line = obj[0]
        close_line = obj[-1]
        obj_inner = obj[1:-1]
        obj_indent = mm.group("indent")
        obj_child_indent = obj_indent + "  "
        patched_inner, did = _json_patch_or_create_tags_in_obj(obj_inner, obj_child_indent)
        if did:
            changed = True
        out.extend([open_line] + patched_inner + [close_line])
        i = obj_end

    new_window = window[:rep_open+1] + out + window[rep_end:]
    return new_window, changed

def patch_json(path: str) -> bool:
    try:
        with open(path, "r", encoding="utf-8", newline=None) as fh:
            raw = fh.read()
            nl_seen = fh.newlines
        lines = raw.splitlines()
        had_nl = _had_final_newline(raw)
    except Exception:
        print(f"Skip (unreadable JSON): {path}")
        return False

    out, i, changed = [], 0, False
    while i < len(lines):
        m_props = JSON_PROPS_OPEN_RE.match(lines[i])
        if not m_props:
            out.append(lines[i]); i += 1; continue

        props_indent = m_props.group("indent")
        props_end = _json_obj_end(lines, i)
        if props_end is None:
            out.append(lines[i]); i += 1; continue

        rtype = _json_nearest_type(lines, i, props_indent, props_end)
        window = lines[i+1:props_end]

        child_indent = props_indent + "  "
        # For GlobalTable, the Tags live under each replica object
        if rtype == "AWS::DynamoDB::GlobalTable":
            new_window, did_change = _json_patch_replicas_tags(window, child_indent)
        else:
            new_window, did_change = _json_patch_or_create_tags(window, child_indent)

        out.append(lines[i]); out.extend(new_window); out.append(lines[props_end])
        changed |= did_change
        i = props_end + 1

    if changed:
        write_nl = _pick_write_newline(nl_seen)
        tail = "\n" if had_nl else ""
        with open(path, "w", encoding="utf-8", newline=write_nl) as fh:
            fh.write("\n".join(out) + tail)
        print(f"Patched (JSON): {path}")
    else:
        print(f"No data storage definitions found (JSON): {path}")
    return changed

def main(paths):
    any_changed = False
    for p in paths:
        lp = p.lower()
        if "test" in lp:
            print(f'Skip (path contains "test"): {p}')
            continue
        if lp.endswith((".yml", ".yaml")):
            any_changed |= patch_yaml(p)
        elif lp.endswith(".json"):
            any_changed |= patch_json(p)
        else:
            print(f"Skip (unknown type): {p}")
    return 0


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("usage: cf_script.py <file1> [file2 ...]", file=sys.stderr)
        sys.exit(2)
    sys.exit(main(sys.argv[1:]))
