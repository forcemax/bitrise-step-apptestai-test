# apptestai-test

bitrise step for apptest.ai test execution

## How to use this Step

Input your apptest.ai Access Key, Project ID, Package file <br />
refer to more information from https://app.apptest.ai/#/main/integrations

**Required** apptest.ai Access Key, Project ID, Package file.

We recommend to use Access Key using bitrise workflow secret variable.

### Example usage
This is the example to using bitrise workflow bitrise.yml<br />
Please change to the your input.

```yaml
    - apptestai-test@0.0.3:
      inputs:
        - access_key: "ci_support:4b55ec5999ea636b0aafb402816ac50b"
        - project_id: "1111"
        - binary_path: "app/build/outputs/apk/prod/release/app-prod-release.apk"
```
