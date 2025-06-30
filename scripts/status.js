#!/usr/bin/env node

const { program } = require('commander');
const Utils = require('./utils');
const http = require('http');

class StatusChecker {
  async check() {
    console.log('==========================================');
    console.log('         RecoNya Service Status           ');
    console.log('==========================================\n');

    try {
      // Check directory and get project root
      const projectRoot = Utils.validateRecoNyaDirectory();

      // Check dependencies
      await this.checkDependencies();

      // Check services
      await this.checkServices();

      // Check configuration
      await this.checkConfiguration(projectRoot);

    } catch (error) {
      Utils.log.error('Status check failed: ' + error.message);
      process.exit(1);
    }
  }

  async checkDependencies() {
    Utils.log.info('Checking dependencies...');

    const deps = [
      { name: 'Go', command: 'go', versionFlag: 'version' },
      { name: 'Node.js', command: 'node', versionFlag: '--version' },
      { name: 'npm', command: 'npm', versionFlag: '--version' },
      { name: 'nmap', command: 'nmap', versionFlag: '--version' }
    ];

    for (const dep of deps) {
      if (Utils.commandExists(dep.command)) {
        try {
          const { stdout } = await Utils.runCommandWithOutput(dep.command, [dep.versionFlag]);
          const version = this.extractVersion(stdout);
          Utils.log.success(`${dep.name}: ${version}`);
        } catch {
          Utils.log.warning(`${dep.name}: installed but version check failed`);
        }
      } else {
        Utils.log.error(`${dep.name}: not installed`);
      }
    }
    console.log();
  }

  async checkServices() {
    Utils.log.info('Checking backend service...');

    // Check backend (port 3008)
    const backendProcess = await Utils.findProcessByPort(3008);
    if (backendProcess) {
      Utils.log.success(`Backend: running (PID ${backendProcess.pid})`);
      
      // Test backend web interface
      try {
        const webWorking = await this.testHTTP('http://localhost:3008');
        if (webWorking) {
          Utils.log.success('Backend web interface: accessible');
        } else {
          Utils.log.warning('Backend web interface: not accessible');
        }
      } catch {
        Utils.log.warning('Backend web interface: connection failed');
      }

      // Test backend API
      try {
        const apiWorking = await this.testAPI('http://localhost:3008/api/system-status');
        if (apiWorking) {
          Utils.log.success('Backend API: responding');
        } else {
          Utils.log.warning('Backend API: not responding');
        }
      } catch {
        Utils.log.warning('Backend API: connection failed');
      }
    } else {
      Utils.log.error('Backend: not running');
    }

    console.log();
  }

  async checkConfiguration(projectRoot) {
    Utils.log.info('Checking configuration...');

    const fs = require('fs');
    const path = require('path');

    // Check .env file
    const envPath = path.join(projectRoot, 'backend', '.env');
    if (fs.existsSync(envPath)) {
      Utils.log.success('Backend .env: exists');
      
      // Parse and validate env
      try {
        const envContent = fs.readFileSync(envPath, 'utf8');
        const envVars = this.parseEnvFile(envContent);
        
        if (envVars.NETWORK_RANGE) {
          Utils.log.success(`Network range: ${envVars.NETWORK_RANGE}`);
        } else {
          Utils.log.warning('Network range: not configured');
        }
        
        if (envVars.LOGIN_USERNAME) {
          Utils.log.success(`Login username: ${envVars.LOGIN_USERNAME}`);
        } else {
          Utils.log.warning('Login username: not configured');
        }
      } catch {
        Utils.log.warning('Backend .env: parse error');
      }
    } else {
      Utils.log.error('Backend .env: missing');
    }

    // Check backend dependencies
    const backendModules = path.join(projectRoot, 'backend', 'go.mod');
    
    if (fs.existsSync(backendModules)) {
      Utils.log.success('Backend dependencies: go.mod exists');
    } else {
      Utils.log.error('Backend dependencies: go.mod missing');
    }

    // Check if templates exist
    const templatesDir = path.join(projectRoot, 'backend', 'templates');
    if (fs.existsSync(templatesDir)) {
      Utils.log.success('HTMX templates: directory exists');
    } else {
      Utils.log.error('HTMX templates: directory missing');
    }

    console.log();
  }

  extractVersion(output) {
    const lines = output.split('\n');
    for (const line of lines) {
      const match = line.match(/(\d+\.\d+(?:\.\d+)?)/);
      if (match) {
        return match[1];
      }
    }
    return 'unknown';
  }

  parseEnvFile(content) {
    const vars = {};
    const lines = content.split('\n');
    
    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed && !trimmed.startsWith('#')) {
        const [key, ...valueParts] = trimmed.split('=');
        if (key && valueParts.length > 0) {
          vars[key.trim()] = valueParts.join('=').replace(/"/g, '');
        }
      }
    }
    
    return vars;
  }

  async testAPI(url) {
    return new Promise((resolve) => {
      const req = http.get(url, { timeout: 5000 }, (res) => {
        resolve(res.statusCode === 200);
      });
      
      req.on('error', () => resolve(false));
      req.on('timeout', () => {
        req.destroy();
        resolve(false);
      });
    });
  }

  async testHTTP(url) {
    return new Promise((resolve) => {
      const req = http.get(url, { timeout: 5000 }, (res) => {
        resolve(res.statusCode < 400);
      });
      
      req.on('error', () => resolve(false));
      req.on('timeout', () => {
        req.destroy();
        resolve(false);
      });
    });
  }
}

// Main execution
if (require.main === module) {
  program
    .name('reconya-status')
    .description('Check RecoNya service status')
    .version('1.0.0');

  program.parse();

  const statusChecker = new StatusChecker();
  statusChecker.check().catch(error => {
    Utils.log.error('Status check failed: ' + error.message);
    process.exit(1);
  });
}

module.exports = StatusChecker;