# AMAG - Azure Metrics Aggregator

Amag is a command-line tool for aggregating metrics from Kusto Query Language (KQL) files and saving them as custom metrics or logs in Azure Monitor. It allows you to run KQL queries against Azure Log Analytics workspaces and transform the results into metrics or logs that can be monitored and visualized over time.

## Installation

To install amag, use the `go install` command:

```bash
go install github.com/DrBushyTop/amag@latest
```

Make sure that your GOPATH/bin is added to your PATH environment variable so that you can run amag from the command line.

## Configuration

Amag can be configured using command-line flags, environment variables, or a configuration file. You can set default values using the config commands.

## Authentication

Amag uses Azure Identity for authentication, leveraging `DefaultAzureCredential`. Ensure you're authenticated with Azure CLI or Azure PowerShell before running amag commands.

**Example:**

```bash
az login
```

## Commands and Usage

### 1. Aggregate Metric Command

Aggregate KQL query results and save them as [Custom Azure Monitor metrics](https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-custom-overview).

**Pre-requisites:**
- You will need a pre-existing log analytics workspace.
- You will need to have the Monitoring Metrics Publisher role assigned to the user or service principal running the tool.

**Usage:**

```bash
amag aggregate metric --file /path/to/query.kql --metric LatencyP90 --workspaceid <workspace-id> --scoperesourceid <scope-resource-id>
```

**Example:**

```bash
amag aggregate metric --file ./queries/latency_p90.kql --metric LatencyP90 --workspaceid "12345678-1234-1234-1234-123456789abc" --scoperesourceid "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/MyResourceGroup/providers/Microsoft.Compute/virtualMachines/MyVM"
```

### 2. Aggregate Log Command

Aggregate KQL query results and save them as [custom logs in Azure Monitor Log Analytics Workspace](https://learn.microsoft.com/en-us/azure/azure-monitor/logs/logs-ingestion-api-overview).

**Pre-requisites:**
- You will need a pre-existing log analytics workspace.
- You will also need to have the data collection rule, endpoint and stream name for the custom log. These can be created by running bicep file `./lawsetup/main.bicep` in this repository. It will also output the required params for the command you can use.
    - The bicep file will also set up the required permissions for given user to write custom logs. These can take a while to propagate. 
    - ```powershell
      New-AzResourceGroupDeployment -ResourceGroupName "amag" -TemplateFile "./lawsetup/main.bicep" -logAnalyticsName "amagws" -location "swedencentral" -metricsPublisherObjectId "f8c353f7-0b21-4a15-9693-a9ebf5f5c073" -dataCollectionEndpointName "amagendpoint" -dataCollectionRuleName "amagrule" -enableErrorDiagnosticLogs $true
      ```

**Usage:**

```bash
amag aggregate log --file "/path/to/query.kql" --metric "LatencyP90" --workspaceid "<workspace-id>" --datacollectionendpoint "<data-collection-endpoint>" --datacollectionstreamname "<data-collection-stream-name>" --datacollectionruleid "<data-collection-rule-id>"
```

**Example:**

```bash
amag aggregate log --file ./queries/latency_p90.kql --metric LatencyP90 --workspaceid "12345678-1234-1234-1234-123456789abc" --datacollectionendpoint "https://dc.applicationinsights.azure.com/" --datacollectionstreamname "CustomLogStream" --datacollectionruleid "dcr-12345678-1234-1234-1234-123456789abc"
```


### 3. Config Commands

#### a. Set Configuration Value

Set a default value for a configuration key.

**Usage:**

```bash
amag config set [key] [value]
```

**Example:**

```bash
amag config set workspaceid "12345678-1234-1234-1234-123456789abc"
```

#### b. Show Current Configuration

Display the current configuration settings.

**Usage:**

```bash
amag config show
```

**Example:**

```bash
amag config show
```

**Output:**

```bash
Current configuration:
workspaceid: 12345678-1234-1234-1234-123456789abc
scoperesourceid: /subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/MyResourceGroup/providers/Microsoft.Compute/virtualMachines/MyVM
```

#### c. Load Configuration from a File

Load configuration settings from a YAML file. This will overwrite any existing configuration settings in the default configuration file.

**Usage:**

```bash
amag config load [path]
```

**Example:**

```bash
amag config load ./amag_config.yaml
```

`amag_config.yaml`:

```yaml
workspaceid: 12345678-1234-1234-1234-123456789abc
scoperesourceid: /subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/MyResourceGroup/providers/Microsoft.Compute/virtualMachines/MyVM
datacollectionendpoint: https://dc.applicationinsights.azure.com/
datacollectionstreamname: CustomLogStream
datacollectionruleid: dcr-12345678-1234-1234-1234-123456789abc
```

Note: After loading the configuration file, you can use the `amag config show` command to verify the settings.

### 4. Using a Custom Configuration File

By default, amag looks for a configuration file in `$HOME/.amag/config.yaml`. You can specify a custom configuration file using the `--config` flag with any command.

**Usage:**

```bash
amag --config /path/to/custom_config.yaml aggregate metric --file /path/to/query.kql --metric LatencyP90
```

**Example:**

```bash
amag --config ./my_custom_config.yaml aggregate metric --file ./queries/latency_p90.kql --metric LatencyP90
```

`my_custom_config.yaml`:

```yaml
workspaceid: 12345678-1234-1234-1234-123456789abc
scoperesourceid: /subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/MyResourceGroup/providers/Microsoft.Compute/virtualMachines/MyVM
```

Note: The custom configuration file allows you to override default settings or provide environment-specific configurations.


## Help and Support

For more information on any command, use the `--help` flag:

```bash
amag aggregate metric --help
```

## Contributing

Contributions are welcome! Please submit a pull request or open an issue on GitHub.