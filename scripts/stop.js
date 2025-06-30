#!/usr/bin/env node

const { program } = require('commander');
const Utils = require('./utils');

class ServiceStopper {
  async stop() {
    console.log('==========================================');
    console.log('         Stopping RecoNya Backend        ');
    console.log('==========================================\n');

    try {
      let stoppedAny = false;

      // Stop backend (port 3008)
      const backendStopped = await Utils.killProcessByPort(3008, 'backend');
      if (backendStopped) stoppedAny = true;

      // Also try to kill any Go processes that might be RecoNya
      await this.killRecoNyaProcesses();

      if (stoppedAny) {
        Utils.log.success('RecoNya backend stopped');
      } else {
        Utils.log.info('No RecoNya backend was running');
      }

    } catch (error) {
      Utils.log.error('Failed to stop backend: ' + error.message);
      process.exit(1);
    }
  }

  async killRecoNyaProcesses() {
    try {
      // Kill any processes that might be RecoNya backend
      if (Utils.isWindows()) {
        // Windows: Find and kill Go processes running RecoNya
        const { stdout } = await Utils.runCommandWithOutput('tasklist', ['/FI', 'IMAGENAME eq go.exe']);
        const lines = stdout.split('\n');
        
        for (const line of lines) {
          if (line.includes('go.exe')) {
            const parts = line.split(/\s+/);
            const pid = parts[1];
            if (pid && !isNaN(pid)) {
              await Utils.killProcess(parseInt(pid), true);
              Utils.log.info(`Killed Go process ${pid}`);
            }
          }
        }
      } else {
        // Unix-like: Kill processes by name pattern
        try {
          await Utils.runCommand('pkill', ['-f', 'go run.*cmd'], { silent: true });
        } catch {
          // Ignore errors - process might not exist
        }

      }
    } catch (error) {
      // Ignore errors in this cleanup phase
    }
  }
}

// Main execution
if (require.main === module) {
  program
    .name('reconya-stop')
    .description('Stop RecoNya services')
    .version('1.0.0');

  program.parse();

  const stopper = new ServiceStopper();
  stopper.stop().catch(error => {
    Utils.log.error('Failed to stop services: ' + error.message);
    process.exit(1);
  });
}

module.exports = ServiceStopper;