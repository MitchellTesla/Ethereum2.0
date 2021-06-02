// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

export function getDefinitionTemplate() {
  const properties: { [key: string]: any; } = {};

  return {
    // tslint:disable-next-line:max-line-length
    $schema: 'https://schema.management.azure.com/providers/Microsoft.Logic/schemas/2016-06-01/workflowdefinition.json#',
    actions: {},
    contentVersion: '1.0.0.0',
    outputs: {},
    parameters: {
      $connections: {
        defaultValue: {},
        type: 'Object',
      },
    },
    triggers: {
      manual: {
        inputs: {
          schema: {
            properties,
            type: 'object',
          },
        },
        kind: 'Http',
        type: 'Request',
      },
    },
  };
}
