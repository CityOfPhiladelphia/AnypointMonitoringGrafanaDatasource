import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> { }

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { secureJsonFields, secureJsonData } = options;

  const onClientIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        clientId: event.target.value,
      },
    });
  };

  const onClientSecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        clientSecret: event.target.value,
      },
    });
  };

  const onResetClientId = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        clientId: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        clientId: ''
      },
    });
  };

  const onResetClientSecret = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        clientSecret: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        clientSecret: '',
      },
    });
  };

  return (
    <>
      <InlineField label="Client ID" labelWidth={18} interactive tooltip={'Secure field (backend only)'}>
        <SecretInput
          required
          id="config-editor-client-id"
          isConfigured={secureJsonFields.clientId}
          value={secureJsonData?.clientId}
          placeholder="Enter your Client ID"
          width={40}
          onReset={onResetClientId}
          onChange={onClientIdChange}
        />
      </InlineField>
      <InlineField label="Client Secret" labelWidth={18} interactive tooltip={'Secure field (backend only)'}>
        <SecretInput
          required
          id="config-editor-client-secret"
          isConfigured={secureJsonFields.clientSecret}
          value={secureJsonData?.clientSecret}
          placeholder="Enter your Client Secret"
          width={40}
          onReset={onResetClientSecret}
          onChange={onClientSecretChange}
        />
      </InlineField>
    </>
  );
}
