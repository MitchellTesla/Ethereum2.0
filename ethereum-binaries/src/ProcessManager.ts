import { ChildProcess, spawn } from "child_process";
import { ManagedProcess } from "./types";

export interface SpawnOptions {
  stdio?: 'pipe' | 'inherit',
  timeout?: number
}

export class ProcessManager {
  private _processes: Array<ManagedProcess>
  constructor() {
    this._processes = []
  }
  add(process: ChildProcess, clientId: string) {
    this._processes.push({
      processId: `${process.pid}`,
      process,
      clientId
    })
  }
  /**
   * wait for process to finish or kill after timeout
   * @param proc 
   * @param timeout 
   */
  onExit(proc: ChildProcess, timeout? : number) {
    return new Promise((resolve, reject) => {
      const cleanup = () => {
        proc.off('exit', onExit)
        proc.off('close', onExit)
        if (timeoutHandler) {
          clearTimeout(timeoutHandler)
        }
      }
      const onExit = (code: number) => {
        cleanup()
        resolve(code)
      }

      let timeoutHandler : any = undefined
      proc.on('exit', onExit)
      proc.on('close', onExit)
      if (timeout !== undefined) {
        timeoutHandler = setTimeout(() => {
          console.log('timeout reached')
          cleanup()
          this.kill(proc.pid)
          reject(new Error('Command timeout reached: '+timeout))
        }, timeout)
      }
    })
  }
  spawn(clientId: string, command: string, args: ReadonlyArray<string> = [], {
    stdio = 'pipe'
  } : SpawnOptions = {}) {
    const _process = spawn(command, args, {
      // we "simulate" inherit to be able to intercept stdout
      // https://github.com/sindresorhus/execa/issues/121
      // https://github.com/nodejs/node/issues/8033
      stdio: stdio === 'inherit' ? ['inherit', 'pipe', 'pipe'] : [stdio, stdio, stdio],
      detached: false,
      shell: false,
    })
    if (stdio === 'inherit') {
      const { stdout, stderr} = _process
      // please node that this is not a full replacement for 'inherit'
      // the child process can and will detect if stdout is a pty and change output based on it
      // the terminal context is lost & ansi information (coloring) etc will be lost
      // https://stackoverflow.com/questions/1401002/how-to-trick-an-application-into-thinking-its-stdout-is-a-terminal-not-a-pipe
      if (stdout && stderr) {
        stdout.pipe(process.stdout)
        stderr.pipe(process.stderr)
      }
    }
    this.add(_process, clientId)
    return _process
  }
  async exec(clientId: string, command: string, args: ReadonlyArray<string> = [], options? : SpawnOptions) : Promise<ChildProcess> {
    return this.spawn(clientId, command, args, options)
  }
  kill(processId: number | string) {
    if (typeof processId !== 'string') {
      processId = `${processId}`
    }
    const managedProcess = this._processes.find(p => p.processId === processId);
    if (!managedProcess) {
      throw new Error(`Process could not be killed - not found: ${processId}`);
    }
    const killResult = managedProcess.process.kill("SIGKILL")
    // TODO log kill result
    this._processes = this._processes.filter(p => p.processId !== processId);
  }
}