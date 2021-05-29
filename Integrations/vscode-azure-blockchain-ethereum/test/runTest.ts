import * as path from 'path';

import { runTests } from 'vscode-test';

async function main() {
  try {
    const extensionDevelopmentPath = path.resolve(__dirname, '../../');
    const extensionTestsPath = path.resolve(__dirname);

    await runTests({ extensionDevelopmentPath, extensionTestsPath });
  } catch (err) {
    // tslint:disable-next-line: no-console
    console.error('Failed to run tests');
    process.exit(1);
  }
}

main();
