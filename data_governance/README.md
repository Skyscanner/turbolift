# TurboLift Campaign: Data Governance

## Description

This PR is part of a Skyscanner-wide Turbolift campaign to apply baseline Data Governance tags to critical AWS data stores.  
The campaign supports the **2025 data governance objective** and helps address cybersecurity risks by ensuring consistent tagging and clear ownership of data resources.

## Changes

This PR updates AWS data storage resources defined in code:

- Amazon S3 buckets
- Amazon DynamoDB tables
- Amazon RDS databases (`DBInstance`, `DBCluster`)

The following tags have been added to the resource definitions:

```bash
Properties:
    Tags:
        - Key: data_classification
          Value: 'CHANGE_ME'
        - Key: data_category
          Value: 'CHANGE_ME'
```

### Tag Values

Choose from the following values for each tag:

- **`data_classification`**:
  - `public`
      - Safe to share externally, no risk to Skyscanner. Includes public content and marketing materials.
  - `internal`
      - For Skyscanner employees/contractors only, low risk if disclosed. Examples: internal docs or configs.
  - `confidential`
      - Limited to specific teams, disclosure could harm operations or compliance. Examples: partner terms, service logs.
  - `restricted`
      - Highly sensitive, strict access control required. Disclosure could cause legal or financial harm. Examples: traveller or payroll data.


- **`data_category`**:
  - `inventory_data`
      - Core service data Skyscanner defines and manages, representing business objects used to provide travel search results. Examples: flight timetables, quotes, hotel inventory.
  - `service_internal_data`
      - Configuration or build-time data that supports service operation but is not exposed externally (e.g., feature toggles, deployment metadata, service configurations).
  - `traveller_profile_data`
      - Data describing a traveller’s preferences or characteristics (language, currency, market, consent choices).
  - `service_analytical_data`
      - Data about service performance, reliability, and operational excellence. Includes logs, metrics, traces, and cost tracking.
  - `business_analytical_data`
      - Data used to understand business performance and drive decision-making. Examples: revenue, conversion rate, ABPC metrics.
  - `user_behaviour_data`
      - Data describing what users see, click, or interact with. It requires consent and is used to improve user experience.
  - `service_snapshot_data`
      - Captures the state of a service or inventory at a specific time. Used for debugging, auditing, testing, or training models.

### Action Required

Please review the changes in this PR and replace the placeholder `CHANGE_ME` values for `data_classification` and `data_category` with the correct ones for your resource.

Guidance on selecting the right values can be found here:
- [Baseline Data Governance for critical AWS Operational Data Stores](https://skyscanner.atlassian.net/browse/DATAGOV-239)
- [Data Classification Framework](https://skyscanner.atlassian.net/wiki/spaces/GOV/pages/22516568/Data+Classification+Framework)
- [Skyscanner Data Categorisation](https://skyscanner.atlassian.net/wiki/spaces/GOV/pages/103072170/Skyscanner+Data+Categorisation)

When you have chosen the appropriate values, update the tags. For example:

```bash
Properties:
    Tags:
        - Key: data_classification
          Value: 'internal'
        - Key: data_category
          Value: 'service_analytical_data'
```

Teams may optionally add `data_subcategory` tag in a case more than 1 `data_category` applies:

```bash
Properties:
    Tags:
        - Key: data_classification
          Value: 'internal'
        - Key: data_category
          Value: 'service_analytical_data'
        - Key: data_subcategory
          Value: 'business_analytical_data'
```

### Tracking
You can monitor your team’s progress in this campaign through the [Data Governance Turbolift Campaign GitHub Project](https://github.com/orgs/Skyscanner/projects/13).
This project allows you to track all open PRs associated with your team and view their current status.
To narrow down the view to your team’s PRs, use the search filter: `owner:<your-team-name>`.

In addition, you can track the number of tagged resources using [Wiz](https://app.wiz.io/inventory/cloud-resources#%7E%28filters%7E%28cloudPlatform%7E%28equals%7E%28%7E%27AWS%29%29%7Etechnology%7E%28equals%7E%28%7E%273032%7E%273031%7E%277037%7E%276870%7E%277038%7E%277042%7E%277045%7E%2711259%7E%273038%7E%276871%7E%277041%7E%276873%7E%277040%7E%2710014%7E%276874%7E%277044%7E%276872%7E%2710015%7E%273136%29%29%7Etag%7E%28TAG_CONTAINS_ALL%7E%28%7E%28key%7E%27data_category%29%7E%28key%7E%27data_classification%29%29%29%29%7Eentity%7E%28%7E%29%29)
and selecting your Tribe or Team in the `Top Projects` filters.

---

If you are unsure which values to apply, please consult [#data-governance-for-aws-operational-data-stores](https://skyscanner.slack.com/archives/C09GZ0MKKPF).

<sub>This PR was generated using [turbolift](https://github.com/Skyscanner/turbolift).</sub>
