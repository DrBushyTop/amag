param location string
param logAnalyticsName string
param dataCollectionEndpointName string
param dataCollectionRuleName string

@description('Should the diagnostics settings be enabled to send error logs to the Log Analytics workspace')
param enableErrorDiagnosticLogs bool

param metricsPublisherObjectId string = ''

param retentionInDays int = 730
param totalRetentionInDays int = 2556

@description('Short name for the table, used for the stream name and table name. Should not contain the _CL ending. The template will handle that.')
param tableShortName string = 'Aggregates'
var realTableName = '${tableShortName}_CL'
var dataCollectionStreamName = 'Custom-${customTable.name}'

var tableSchema = [
  {
    name: 'TimeGenerated'
    type: 'datetime'
  }
  {
    name: 'OriginalTimeGenerated'
    type: 'datetime'
  }
  {
    name: 'Name'
    type: 'string'
  }
  {
    name: 'Value'
    type: 'real'
  }
]

resource logAnalytics 'Microsoft.OperationalInsights/workspaces@2023-09-01' existing = {
  name: logAnalyticsName
}

resource customTable 'Microsoft.OperationalInsights/workspaces/tables@2022-10-01' = {
  name: realTableName
  parent: logAnalytics
  properties: {
    plan: 'Analytics'
    retentionInDays: retentionInDays
    totalRetentionInDays: totalRetentionInDays
    schema: {
      name: realTableName
      columns: tableSchema
    }
  }
}

resource dataCollectionEndpoint 'Microsoft.Insights/dataCollectionEndpoints@2023-03-11' = {
  name: dataCollectionEndpointName
  location: location
  properties: {
    networkAcls: {
      publicNetworkAccess: 'Enabled'
    }
  }
}

resource dataCollectionRule 'Microsoft.Insights/dataCollectionRules@2023-03-11' = {
  name: dataCollectionRuleName
  location: location
  properties: {
    destinations: {
      logAnalytics: [
        {
          workspaceResourceId: logAnalytics.id
          name: guid(logAnalytics.id)
        }
      ]
    }
    dataCollectionEndpointId: dataCollectionEndpoint.id
    dataFlows: [
      {
        streams: [
          dataCollectionStreamName
        ]
        destinations: [
          guid(logAnalytics.id)
        ]
        outputStream: dataCollectionStreamName
        transformKql: 'source'
      }
    ]
    streamDeclarations: {
      '${dataCollectionStreamName}': {
        columns: tableSchema
      }
    }
  }
}

resource diagnosticsSettings 'Microsoft.Insights/diagnosticSettings@2021-05-01-preview' = if (enableErrorDiagnosticLogs) {
  name: 'logErrorsToLogAnalytics'
  scope: dataCollectionRule
  properties: {
    logs: [
      {
        category: 'LogErrors'
        enabled: true
      }
    ]
    workspaceId: logAnalytics.id
  }
}

resource dataCollectionRulePublisherGroup 'Microsoft.Authorization/roleAssignments@2020-04-01-preview' = if (metricsPublisherObjectId != '') {
  name: guid(metricsPublisherObjectId, dataCollectionEndpoint.id)
  scope: dataCollectionRule
  properties: {
    principalId: metricsPublisherObjectId
    // Monitoring Metrics Publisher
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      '3913510d-42f4-4e42-8a64-420c390055eb'
    )
  }
}

output dataCollectionEndpoint string = dataCollectionEndpoint.properties.logsIngestion.endpoint
output dataCollectionRuleId string = dataCollectionRule.properties.immutableId
output dataCollectionStreamName string = dataCollectionStreamName
output warning string = metricsPublisherObjectId == ''
  ? 'No Monitoring Metrics Publisher permissions set. You should run this deployment again with the metricsPublisherObjectId parameter or set them up manually'
  : ''
