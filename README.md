# apptestai-test

bitrise step for apptest.ai test execution

## How to use this Step

Input your apptest.ai Access Key, Project ID, Package file <br />
refer to more information from https://app.apptest.ai/#/main/integrations

**Required** apptest.ai Access Key, Project ID, Package file.

Setup Access Key using bitrise workflow secret variable : APPTEST_AI_ACCESS_KEY

### Example usage
This is the example to using bitrise workflow bitrise.yml<br />
Please change to the your input.

```yaml
    - apptestai-test@0.0.1:
      inputs:
        - access_key: "$APPTEST_AI_ACCESS_KEY"
        - project_id: "1111"
        - binary_path: "app/build/outputs/apk/prod/release/app-prod-release.apk"
```
