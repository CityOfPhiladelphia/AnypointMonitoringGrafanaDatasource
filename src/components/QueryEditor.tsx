import React, { ChangeEvent } from 'react';
import { InlineField, Input, Stack } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onOrgIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, orgId: event.target.value });
  };
  const onEnvIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, envId: event.target.value });
  };
  const onClusterIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, clusterId: event.target.value });
  };
  const onAppIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, appId: event.target.value });
    // onRunQuery();
  };
  const onMetricNameChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, metricName: event.target.value });
  };
  const onMetricTableChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, metricTable: event.target.value });
  };
  const onTimeStepChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, timeStep: event.target.value });
  };

  const { orgId, envId, clusterId, appId, metricName, metricTable, timeStep } = query;

  return (
    <Stack direction="column" gap={2}>
      <Stack direction="row" gap={2}>
        <InlineField label="Org ID" labelWidth={16} tooltip="Organization ID">
          <Input
            id="query-editor-org-id"
            onChange={onOrgIdChange}
            value={orgId || ''}
            required
            placeholder="Organization ID"
          />
        </InlineField>
        <InlineField label="Env ID" labelWidth={16} tooltip="Environment ID">
          <Input
            id="query-editor-env-id"
            onChange={onEnvIdChange}
            value={envId || ''}
            required
            placeholder="Environment ID"
          />
        </InlineField>
      </Stack>
      <Stack direction="row" gap={2}>
        <InlineField label="Cluster ID" labelWidth={16} tooltip="Cluster ID">
          <Input
            id="query-editor-cluster-id"
            onChange={onClusterIdChange}
            value={clusterId || ''}
            required
            placeholder="Cluster ID"
          />
        </InlineField>
        <InlineField label="App ID" labelWidth={16} tooltip="App ID">
          <Input
            id="query-editor-app-id"
            onChange={onAppIdChange}
            value={appId || ''}
            required
            placeholder="App ID"
          />
        </InlineField>
      </Stack>
      <Stack direction="row" gap={2}>
        <InlineField label="Metric Name" labelWidth={16} tooltip="Metric Name">
          <Input
            id="query-editor-metric-name"
            onChange={onMetricNameChange}
            value={metricName || ''}
            required
            placeholder="Metric Name"
          />
        </InlineField>
        <InlineField label="Metric Table" labelWidth={16} tooltip="Metric Table">
          <Input
            id="query-editor-metric-table"
            onChange={onMetricTableChange}
            value={metricTable || ''}
            required
            placeholder="Metric Table"
          />
        </InlineField>
      </Stack>
      <Stack direction="row" gap={2}>
        <InlineField label="Time Step" labelWidth={16} tooltip="Time Step (Aggregration)">
          <Input
            id="query-editor-time-step"
            onChange={onTimeStepChange}
            value={timeStep || ''}
            required
            placeholder="Time Step"
          />
        </InlineField>
      </Stack>
    </Stack>
  );
}
