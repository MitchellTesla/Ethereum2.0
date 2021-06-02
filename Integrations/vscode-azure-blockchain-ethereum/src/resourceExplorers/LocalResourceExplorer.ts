// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

import { Constants } from '../Constants';
import { showInputBox } from '../helpers';
import { LocalNetworkNode, LocalProject } from '../Models/TreeItems';
import { GanacheService } from '../services';
import { Telemetry } from '../TelemetryClient';
import { DialogResultValidator } from '../validators/DialogResultValidator';
import { UrlValidator } from '../validators/UrlValidator';

export class LocalResourceExplorer {
  public async createProject(existingProjects: string[] = [], existingPorts: number[] = []): Promise<LocalProject> {
    Telemetry.sendEvent('LocalResourceExplorer.createProject');
    return this.getOrCreateLocalProject(
      existingProjects,
      existingPorts,
      GanacheService.PortStatus.FREE,
      Constants.validationMessages.portAlreadyInUse,
    );
  }

  public async selectProject(existingProjects: string[] = [], existingPorts: number[] = []): Promise<LocalProject> {
    Telemetry.sendEvent('LocalResourceExplorer.selectProject');
    const localProject = await this.getOrCreateLocalProject(
      existingProjects,
      existingPorts,
      GanacheService.PortStatus.GANACHE,
      Constants.validationMessages.portNotInUseGanache,
    );

    await GanacheService.startGanacheServer(localProject.port);

    return localProject;
  }

  private async getOrCreateLocalProject(
    existingProjects: string[],
    existingPorts: number[],
    portStatus: GanacheService.PortStatus,
    validateMessage: string,
  ): Promise<LocalProject> {
    const localProjectName = await this.getLocalProjectName(existingProjects);
    const localProjectPort = await this.getLocalProjectPort(existingPorts, portStatus, validateMessage);

    return this.getLocalProject(localProjectName, localProjectPort);
  }

  private async getLocalProjectName(existingProjects: string[]): Promise<string> {
    return showInputBox({
      ignoreFocusOut: true,
      prompt: Constants.paletteLabels.enterLocalProjectName,
      validateInput: (value: string) => {
        const validationError = DialogResultValidator.validateLocalNetworkName(value);
        if (validationError) {
          return validationError;
        }

        if (existingProjects.some((existName) => existName === value)) {
          return Constants.validationMessages.nameAlreadyInUse;
        }

        return null;
      },
      value: existingProjects.includes(Constants.localhostName) ? '' : Constants.localhostName,
    });
  }

  private async getLocalProjectPort(
    existingPorts: number[],
    portStatus: GanacheService.PortStatus,
    validateMessage: string,
  ): Promise<number> {
    const port = await showInputBox({
      ignoreFocusOut: true,
      prompt: Constants.paletteLabels.enterLocalProjectPort,
      validateInput: async (value: string) => {
        const validationError = UrlValidator.validatePort(value);
        if (validationError) {
          return validationError;
        }

        if (existingPorts.some((existPort) => existPort + '' === value)) {
          return Constants.validationMessages.projectAlreadyExists;
        }

        if (await GanacheService.getPortStatus(value) !== portStatus) {
          return validateMessage;
        }

        return null;
      },
      value: existingPorts.includes(Constants.defaultLocalhostPort) ? '' : Constants.defaultLocalhostPort.toString(),
    });

    return parseInt(port, 10);
  }

  private async getLocalProject(projectName: string, port: number): Promise<LocalProject> {
    const localProject = new LocalProject(projectName, port);
    const url = `${Constants.networkProtocols.http}${Constants.localhost}:${port}`;
    const networkNode = new LocalNetworkNode(projectName, url, '*');

    localProject.addChild(networkNode);

    return localProject;
  }
}
