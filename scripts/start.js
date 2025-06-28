#!/usr/bin/env node

const { program } = require('commander');
const { spawn } = require('child_process');
const chalk = require('chalk');
const Utils = require('./utils');
const path = require('path');

class ServiceManager {
  constructor() {
    this.backendProcess = null;
    this.frontendProcess = null;
    this.isShuttingDown = false;
  }

  async start() {
    console.log('==========================================');
    console.log('         Starting RecoNya Services       ');
    console.log('==========================================\n');

    // Validate directory and get project root
    const projectRoot = Utils.validateRecoNyaDirectory();

    Utils.log.info('Backend will run on: http://localhost:3008');
    Utils.log.info('Frontend will run on: http://localhost:3000\n');

    try {
      // Check and free required ports
      Utils.log.info('Checking for existing processes on required ports...');
      await Utils.killProcessByPort(3008, 'backend');
      await Utils.killProcessByPort(3000, 'frontend');

      Utils.log.info('Press Ctrl+C to stop both services\n');

      // Setup signal handlers
      this.setupSignalHandlers();

      // Start backend
      await this.startBackend();

      // Start frontend
      await this.startFrontend();

      Utils.log.success('RecoNya is starting up...');
      Utils.log.info('Once both services are ready, open your browser to: http://localhost:3000');
      Utils.log.info('Default login: admin / password\n');

      // Keep the process alive
      await this.waitForShutdown();

    } catch (error) {
      Utils.log.error('Failed to start RecoNya: ' + error.message);
      await this.cleanup();
      process.exit(1);
    }
  }

  async startBackend() {
    Utils.log.info('Starting backend...');

    return new Promise((resolve, reject) => {
      const projectRoot = Utils.validateRecoNyaDirectory();
      const backendPath = path.join(projectRoot, 'backend');
      
      this.backendProcess = spawn('go', ['run', './cmd'], {
        cwd: backendPath,
        stdio: ['ignore', 'pipe', 'pipe'],
        shell: Utils.isWindows()
      });

      let startupTimeout;
      let isStarted = false;

      // Handle backend output
      this.backendProcess.stdout.on('data', (data) => {
        const output = data.toString();
        console.log(chalk.gray('[BACKEND]'), output.trim());
        
        // Check if backend is ready
        if (output.includes('Server is starting on port 3008') || 
            output.includes('Starting new ping sweep scan')) {
          if (!isStarted) {
            isStarted = true;
            clearTimeout(startupTimeout);
            Utils.log.success('Backend started successfully');
            resolve();
          }
        }
      });

      this.backendProcess.stderr.on('data', (data) => {
        console.log(chalk.red('[BACKEND ERROR]'), data.toString().trim());
      });

      this.backendProcess.on('close', (code) => {
        if (!this.isShuttingDown && !isStarted) {
          reject(new Error(`Backend exited with code ${code}`));
        }
      });

      this.backendProcess.on('error', (error) => {
        if (!isStarted) {
          reject(error);
        }
      });

      // Set startup timeout
      startupTimeout = setTimeout(() => {
        if (!isStarted) {
          reject(new Error('Backend startup timeout'));
        }
      }, 30000);
    });
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

  async startFrontend() {
    Utils.log.info('Starting frontend...');

    return new Promise((resolve, reject) => {
      const projectRoot = Utils.validateRecoNyaDirectory();
      const frontendPath = path.join(projectRoot, 'frontend');
      
      // Set environment variables for frontend
      const env = {
        ...process.env,
        BROWSER: 'none', // Don't auto-open browser
        CI: 'true' // Reduce output verbosity
      };

      this.frontendProcess = spawn('npm', ['start'], {
        cwd: frontendPath,
        stdio: ['ignore', 'pipe', 'pipe'],
        shell: Utils.isWindows(),
        env
      });

      let startupTimeout;
      let isStarted = false;

      // Handle frontend output
      this.frontendProcess.stdout.on('data', (data) => {
        const output = data.toString();
        
        // Filter out verbose webpack output
        if (output.includes('webpack compiled') || 
            output.includes('Local:') ||
            output.includes('On Your Network:')) {
          console.log(chalk.gray('[FRONTEND]'), output.trim());
          
          if (output.includes('webpack compiled') && !isStarted) {
            isStarted = true;
            clearTimeout(startupTimeout);
            Utils.log.success('Frontend started successfully');
            resolve();
          }
        }
      });

      this.frontendProcess.stderr.on('data', (data) => {
        const output = data.toString();
        // Filter out common React warnings
        if (!output.includes('WARNING in') && !output.includes('Module Warning')) {
          console.log(chalk.yellow('[FRONTEND WARNING]'), output.trim());
        }
      });

      this.frontendProcess.on('close', (code) => {
        if (!this.isShuttingDown && !isStarted) {
          reject(new Error(`Frontend exited with code ${code}`));
        }
      });

      this.frontendProcess.on('error', (error) => {
        if (!isStarted) {
          reject(error);
        }
      });

      // Set startup timeout
      startupTimeout = setTimeout(() => {
        if (!isStarted) {
          Utils.log.warning('Frontend startup timeout, but continuing...');
          resolve();
        }
      }, 60000);
    });
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

    Utils.log.info('Stopping services...');

    const cleanupPromises = [];

    // Kill backend
    if (this.backendProcess && !this.backendProcess.killed) {
      cleanupPromises.push(this.killProcessGracefully(this.backendProcess, 'Backend'));
    }

    // Kill frontend
    if (this.frontendProcess && !this.frontendProcess.killed) {
      cleanupPromises.push(this.killProcessGracefully(this.frontendProcess, 'Frontend'));
    }

    await Promise.all(cleanupPromises);
    Utils.log.success('Services stopped');
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
    .description('Start RecoNya services')
    .version('1.0.0');

  program.parse();

  const serviceManager = new ServiceManager();
  serviceManager.start().catch(error => {
    Utils.log.error('Failed to start services: ' + error.message);
    process.exit(1);
  });
}

module.exports = ServiceManager;