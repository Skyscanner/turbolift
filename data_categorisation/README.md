# TurboLift Campaign: Data Categorisation

## Description

This PR is part of a Skyscanner-wide Turbolift campaign to apply baseline Data Governance tags to critical AWS data stores.  
The campaign supports the **2025 data governance objective** and helps reduce audit risks by ensuring consistent tagging and clear ownership of data resources.

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
          Value: '{{resolve:ssm:/DataClassification/CHANGE_ME}}'
        - Key: data_category
          Value: '{{resolve:ssm:/DataCategory/CHANGE_ME}}'
```

Teams may optionally add more than one `data_category` by comma-separating values, for example:

```bash
Value: '{{resolve:ssm:/DataCategory/value1}},{{resolve:ssm:/DataCategory/value2}}'
```

### Tag Values

The tag values are stored as **SSM parameters** in each AWS account/region. Choose from the following values:

- **`data_classification`** (choose one):
  - `public`
  - `internal`
  - `confidential`
  - `restricted`


- **`data_category`** (choose one or more):
  - `inventory_data`
  - `traveller_profile_data`
  - `service_analytical_data`
  - `business_analytical_data`
  - `user_behaviour_data`
  - `service_snapshot_data`

### Action Required

Please review the resources in this repository and replace the placeholder `CHANGE_ME` values for `data_classification` and `data_category` with the correct ones for your service.

Guidance on selecting the right values can be found here:
- [Data Classification Framework](https://skyscanner.atlassian.net/wiki/spaces/GOV/pages/22516568/Data+Classification+Framework)
- [Skyscanner Data Categorisation](https://skyscanner.atlassian.net/wiki/spaces/GOV/pages/103072170/Skyscanner+Data+Categorisation)

When you have chosen the appropriate values, update the tags. For example:

```bash
Properties:
    Tags:
        - Key: data_classification
          Value: '{{resolve:ssm:/DataClassification/internal}}'
        - Key: data_category
          Value: '{{resolve:ssm:/DataCategory/service_analytical_data}}'
```

- Ensure only **one value** is selected for `data_classification`.
- Select **one or more** appropriate values for `data_category`.

If you are unsure which values to apply, please consult the **Data Governance Slack channel**.

<sub>This PR was generated using [turbolift](https://github.com/Skyscanner/turbolift).</sub>
