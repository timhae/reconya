#!/usr/bin/env node

const { program } = require('commander');
const { spawn } = require('child_process');
const chalk = require('chalk');
const Utils = require('./utils');
const path = require('path');
const fs = require('fs');

class ServiceManager {
  constructor() {
    this.backendProcess = null;
    this.isShuttingDown = false;
  }

  findGoExecutable() {
    const { execSync } = require('child_process');
    const fs = require('fs');
    
    const commonPaths = [
      '/opt/homebrew/bin/go',
      '/usr/local/go/bin/go',
      '/usr/local/bin/go',
      '/usr/bin/go'
    ];
    
    try {
      const goPath = execSync('which go', { encoding: 'utf8' }).trim();
      if (goPath && fs.existsSync(goPath)) {
        return goPath;
      }
    } catch (error) {
      // 'which' command failed, continue with manual search
    }
    
    for (const path of commonPaths) {
      if (fs.existsSync(path)) {
        return path;
      }
    }
    
    return 'go';
  }

  async start() {
    console.log('==========================================');
    console.log('         Starting reconYa Backend        ');
    console.log('==========================================\n');

    // Validate directory and get project root
    const projectRoot = Utils.validatereconYaDirectory();

    // Try to determine the port from .env file
    const envPath = path.join(projectRoot, 'backend', '.env');
    let port = '3008'; // default
    try {
      if (fs.existsSync(envPath)) {
        const envContent = fs.readFileSync(envPath, 'utf8');
        const portMatch = envContent.match(/^PORT=(.+)$/m);
        if (portMatch) {
          port = portMatch[1].trim();
        }
      }
    } catch (error) {
      // Use default port if can't read .env
    }

    Utils.log.info(`reconYa backend will run on: http://localhost:${port}`);
    Utils.log.info('HTMX frontend is served directly from the backend\n');

    try {
      // Check and free required port
      Utils.log.info(`Checking for existing processes on port ${port}...`);
      await Utils.killProcessByPort(parseInt(port), 'backend');

      Utils.log.info('Press Ctrl+C to stop the service\n');

      // Setup signal handlers
      this.setupSignalHandlers();

      // Start backend
      await this.startBackend();

      Utils.log.success('reconYa backend is starting up...');
      Utils.log.info(`Open your browser to: http://localhost:${port}`);
      Utils.log.info('Default login: admin / password\n');

      // Keep the process alive
      await this.waitForShutdown();

    } catch (error) {
      Utils.log.error('Failed to start reconYa: ' + error.message);
      await this.cleanup();
      process.exit(1);
    }
  }

  async startBackend() {
    Utils.log.info('Starting backend with immortal restart capability...');
    
    const projectRoot = Utils.validatereconYaDirectory();
    const backendPath = path.join(projectRoot, 'backend');
    let restartCount = 0;
    let isFirstStart = true;

    const startBackendProcess = () => {
      return new Promise((resolve, reject) => {
        if (!isFirstStart) {
          restartCount++;
          const delay = Math.min(restartCount * 1000, 5000); // Max 5 second delay
          Utils.log.info(`Restarting backend in ${delay}ms (attempt #${restartCount})...`);
          setTimeout(() => {
            this.createBackendProcess(backendPath, resolve, reject);
          }, delay);
        } else {
          isFirstStart = false;
          this.createBackendProcess(backendPath, resolve, reject);
        }
      });
    };

    // Start the backend with auto-restart
    const startWithRestart = async () => {
      while (!this.isShuttingDown) {
        try {
          await startBackendProcess();
          // Reset restart count on successful startup
          if (restartCount > 0) {
            restartCount = 0;
            Utils.log.success('Backend running stable, restart counter reset');
          }
          break;
        } catch (error) {
          if (this.isShuttingDown) break;
          Utils.log.warning(`Backend failed: ${error.message}`);
          Utils.log.info('Backend will restart automatically...');
        }
      }
    };

    startWithRestart();
    
    // Return immediately since we're handling restarts automatically
    return new Promise((resolve) => {
      // Give the first start attempt time to initialize
      setTimeout(() => {
        Utils.log.success('Backend startup initiated with immortal protection');
        resolve();
      }, 2000);
    });
  }

  createBackendProcess(backendPath, resolve, reject) {
    // Ensure Go is in the PATH by checking common locations
    const goPath = this.findGoExecutable();
    if (!goPath) {
      reject(new Error('Go executable not found. Please ensure Go is installed and in your PATH.'));
      return;
    }

    this.backendProcess = spawn(goPath, ['run', './cmd'], {
      cwd: backendPath,
      stdio: ['ignore', 'pipe', 'pipe'],
      shell: Utils.isWindows(),
      env: process.env
    });

    let startupTimeout;
    let isStarted = false;

    // Handle backend output
    this.backendProcess.stdout.on('data', (data) => {
      const output = data.toString();
      console.log(chalk.gray('[BACKEND]'), output.trim());
      
      // Look for multiple startup indicators
      if (output.includes('Backend startup completed successfully') ||
          output.includes('reconYa backend is ready') ||
          output.includes('Server is starting on port 3008') || 
          output.includes('Starting new ping sweep scan')) {
        if (!isStarted) {
          isStarted = true;
          clearTimeout(startupTimeout);
          resolve();
        }
      }
    });

    this.backendProcess.stderr.on('data', (data) => {
      const output = data.toString().trim();
      
      // Only flag actual errors/panics as errors, treat everything else as info
      if (output.includes('panic:') || 
          output.includes('FATAL') ||
          output.includes('Error:') ||
          output.includes('failed:') ||
          output.includes('PANIC') ||
          output.includes('Stack trace:') ||
          output.includes('runtime error') ||
          output.includes('invalid memory address') ||
          output.includes('nil pointer dereference')) {
        // These are actual errors
        console.log(chalk.red('[BACKEND ERROR]'), output);
      } else {
        // Everything else is normal operational logs
        console.log(chalk.gray('[BACKEND]'), output);
      }
    });

    this.backendProcess.on('close', (code) => {
      if (!this.isShuttingDown) {
        Utils.log.warning(`Backend exited with code ${code} - will restart automatically`);
        setTimeout(() => {
          if (!this.isShuttingDown) {
            this.startBackend(); // Auto-restart
          }
        }, 1000);
      }
    });

    this.backendProcess.on('error', (error) => {
      if (!isStarted && !this.isShuttingDown) {
        reject(error);
      }
    });

    // Extended timeout for backend startup (60 seconds)
    startupTimeout = setTimeout(() => {
      if (!isStarted) {
        Utils.log.warning('Backend startup taking longer than expected, but continuing...');
        isStarted = true;
        resolve();
      }
    }, 60000);
  }

  async verifyBackendStartup() {
    Utils.log.info('Verifying backend startup...');

    // Wait for backend to be listening on port 3008
    const isListening = await Utils.waitForPort(3008, 30000);
    
    if (!isListening) {
      throw new Error('Backend is not listening on port 3008');
    }

    Utils.log.success('Backend started successfully');
  }


  setupSignalHandlers() {
    const signals = ['SIGINT', 'SIGTERM', 'SIGQUIT'];
    
    signals.forEach(signal => {
      process.on(signal, async () => {
        if (!this.isShuttingDown) {
          Utils.log.info('\nShutdown signal received...');
          await this.cleanup();
          process.exit(0);
        }
      });
    });

    // Handle unexpected exits
    process.on('exit', () => {
      this.cleanup();
    });

    process.on('uncaughtException', async (error) => {
      Utils.log.error('Uncaught exception: ' + error.message);
      await this.cleanup();
      process.exit(1);
    });
  }

  async cleanup() {
    if (this.isShuttingDown) return;
    this.isShuttingDown = true;

    Utils.log.info('Stopping backend service...');

    // Kill backend
    if (this.backendProcess && !this.backendProcess.killed) {
      await this.killProcessGracefully(this.backendProcess, 'Backend');
    }

    Utils.log.success('Backend service stopped');
  }

  async killProcessGracefully(proc, serviceName) {
    return new Promise((resolve) => {
      let killed = false;
      
      const forceKill = () => {
        if (!killed && !proc.killed) {
          proc.kill('SIGKILL');
          killed = true;
        }
        resolve();
      };

      // Try graceful kill first
      proc.kill('SIGTERM');
      
      // Set timeout for force kill
      const timeout = setTimeout(forceKill, 5000);
      
      proc.on('close', () => {
        if (!killed) {
          killed = true;
          clearTimeout(timeout);
          resolve();
        }
      });
    });
  }

  async waitForShutdown() {
    return new Promise((resolve) => {
      // This will keep the process alive until cleanup is called
      const keepAlive = setInterval(() => {
        if (this.isShuttingDown) {
          clearInterval(keepAlive);
          resolve();
        }
      }, 1000);
    });
  }
}

// Main execution
if (require.main === module) {
  program
    .name('reconya-start')
    .description('Start reconYa services')
    .option('-d, --daemon', 'run as daemon (detached)')
    .version('1.0.0');

  program.parse();
  const options = program.opts();

  if (options.daemon) {
    // Run as daemon
    const { spawn } = require('child_process');
    const path = require('path');
    const fs = require('fs');
    
    console.log('Starting reconYa backend as daemon...');
    
    // Create logs directory if it doesn't exist
    const logsDir = path.join(process.cwd(), 'logs');
    if (!fs.existsSync(logsDir)) {
      fs.mkdirSync(logsDir, { recursive: true });
    }
    
    // Setup log files
    const logFile = path.join(logsDir, 'reconya.log');
    const errorFile = path.join(logsDir, 'reconya.error.log');
    
    const out = fs.openSync(logFile, 'a');
    const err = fs.openSync(errorFile, 'a');
    
    const child = spawn(process.execPath, [__filename], {
      detached: true,
      stdio: ['ignore', out, err]
    });
    
    child.unref();
    console.log(`reconYa daemon started with PID: ${child.pid}`);
    console.log(`Logs: ${logFile}`);
    console.log(`Errors: ${errorFile}`);
    
    // Write PID file for daemon management
    const pidFile = path.join(process.cwd(), '.reconya.pid');
    fs.writeFileSync(pidFile, child.pid.toString());
    
    process.exit(0);
  } else {
    // Run in terminal (persistent)
    const serviceManager = new ServiceManager();
    serviceManager.start().catch(error => {
      Utils.log.error('Failed to start services: ' + error.message);
      process.exit(1);
    });
  }
}

module.exports = ServiceManager;