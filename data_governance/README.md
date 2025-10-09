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
  - `internal`
  - `confidential`
  - `restricted`


- **`data_category`**:
  - `inventory_data`
  - `service_internal_data`
  - `traveller_profile_data`
  - `service_analytical_data`
  - `business_analytical_data`
  - `user_behaviour_data`
  - `service_snapshot_data`

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

If you are unsure which values to apply, please consult [#data-governance-for-aws-operational-data-stores](https://skyscanner.slack.com/archives/C09GZ0MKKPF).

<sub>This PR was generated using [turbolift](https://github.com/Skyscanner/turbolift).</sub>
