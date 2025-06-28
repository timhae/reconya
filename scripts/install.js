#!/usr/bin/env node

const { program } = require('commander');
const inquirer = require('inquirer');
const Utils = require('./utils');
const fs = require('fs');
const path = require('path');

class Installer {
  constructor() {
    this.os = Utils.getOS();
  }

  async run() {
    console.log('==========================================');
    console.log('       RecoNya Installation Script       ');
    console.log('==========================================\n');

    Utils.log.info(`Detected ${this.os}`);

    try {
      await this.installDependencies();
      await this.setupRecoNya();
      await this.createScripts();
      
      console.log('\n==========================================');
      Utils.log.success('RecoNya installation completed!');
      console.log('==========================================\n');
      
      Utils.log.info('To start RecoNya, run:');
      console.log('  npm run start\n');
      Utils.log.info('Or use the individual commands:');
      console.log('  npm run status  - Check service status');
      console.log('  npm run start   - Start RecoNya');
      console.log('  npm run stop    - Stop RecoNya');
      console.log('  npm run install - Reinstall dependencies\n');
      Utils.log.info('Then open your browser to: http://localhost:3000');
      Utils.log.info('Default login: admin / password\n');
      Utils.log.warning('Important: Configure your network range in backend/.env');
      
    } catch (error) {
      Utils.log.error('Installation failed: ' + error.message);
      process.exit(1);
    }
  }

  async installDependencies() {
    Utils.log.info('Installing dependencies...');

    switch (this.os) {
      case 'macos':
        await this.installMacOSDeps();
        break;
      case 'debian':
        await this.installDebianDeps();
        break;
      case 'redhat':
        await this.installRedHatDeps();
        break;
      case 'windows':
        await this.installWindowsDeps();
        break;
      default:
        throw new Error(`Unsupported operating system: ${this.os}`);
    }

    await Utils.setupNmapPermissions();
  }

  async installMacOSDeps() {
    Utils.log.info('Installing dependencies for macOS...');

    // Check Homebrew
    if (!Utils.commandExists('brew')) {
      Utils.log.info('Installing Homebrew...');
      await Utils.runCommand('/bin/bash', ['-c', 
        '"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
      ]);
    }

    // Install Go
    if (!Utils.commandExists('go')) {
      Utils.log.info('Installing Go...');
      await Utils.runCommand('brew', ['install', 'go']);
    } else {
      const { stdout } = await Utils.runCommandWithOutput('go', ['version']);
      Utils.log.success(`Go is already installed (${stdout.trim()})`);
    }

    // Install Node.js
    if (!Utils.commandExists('node')) {
      Utils.log.info('Installing Node.js...');
      await Utils.runCommand('brew', ['install', 'node']);
    } else {
      const { stdout } = await Utils.runCommandWithOutput('node', ['--version']);
      Utils.log.success(`Node.js is already installed (${stdout.trim()})`);
    }

    // Install nmap
    if (!Utils.commandExists('nmap')) {
      Utils.log.info('Installing nmap...');
      await Utils.runCommand('brew', ['install', 'nmap']);
    } else {
      Utils.log.success('nmap is already installed');
    }
  }

  async installDebianDeps() {
    Utils.log.info('Installing dependencies for Debian-based Linux...');

    // Update package list
    Utils.log.info('Updating package list...');
    await Utils.runCommand('sudo', ['apt-get', 'update']);

    // Install basic tools
    await Utils.runCommand('sudo', ['apt-get', 'install', '-y', 
      'curl', 'wget', 'software-properties-common', 'apt-transport-https'
    ]);

    // Install Go
    if (!Utils.commandExists('go')) {
      Utils.log.info('Installing Go...');
      const goVersion = '1.21.5';
      const arch = process.arch === 'arm64' ? 'arm64' : 'amd64';
      
      const { stdout } = await Utils.runCommandWithOutput('wget', [
        `https://golang.org/dl/go${goVersion}.linux-${arch}.tar.gz`,
        '-O', 'go.tar.gz'
      ]);
      
      await Utils.runCommand('sudo', ['rm', '-rf', '/usr/local/go']);
      await Utils.runCommand('sudo', ['tar', '-C', '/usr/local', '-xzf', 'go.tar.gz']);
      await Utils.runCommand('rm', ['go.tar.gz']);

      // Add to PATH
      const bashrc = path.join(process.env.HOME, '.bashrc');
      const pathLine = 'export PATH=$PATH:/usr/local/go/bin';
      
      if (fs.existsSync(bashrc)) {
        const content = fs.readFileSync(bashrc, 'utf8');
        if (!content.includes('/usr/local/go/bin')) {
          fs.appendFileSync(bashrc, `\n${pathLine}\n`);
        }
      }
      
      process.env.PATH += ':/usr/local/go/bin';
    } else {
      const { stdout } = await Utils.runCommandWithOutput('go', ['version']);
      Utils.log.success(`Go is already installed (${stdout.trim()})`);
    }

    // Install Node.js
    if (!Utils.commandExists('node')) {
      Utils.log.info('Installing Node.js...');
      await Utils.runCommand('curl', ['-fsSL', 
        'https://deb.nodesource.com/setup_18.x', '|', 'sudo', '-E', 'bash', '-'
      ]);
      await Utils.runCommand('sudo', ['apt-get', 'install', '-y', 'nodejs']);
    } else {
      const { stdout } = await Utils.runCommandWithOutput('node', ['--version']);
      Utils.log.success(`Node.js is already installed (${stdout.trim()})`);
    }

    // Install nmap
    if (!Utils.commandExists('nmap')) {
      Utils.log.info('Installing nmap...');
      await Utils.runCommand('sudo', ['apt-get', 'install', '-y', 'nmap']);
    } else {
      Utils.log.success('nmap is already installed');
    }
  }

  async installRedHatDeps() {
    Utils.log.info('Installing dependencies for Red Hat-based Linux...');

    const pkgMgr = Utils.commandExists('dnf') ? 'dnf' : 'yum';
    
    // Install basic tools
    await Utils.runCommand('sudo', [pkgMgr, 'install', '-y', 'curl', 'wget']);

    // Install Go (similar to Debian)
    if (!Utils.commandExists('go')) {
      Utils.log.info('Installing Go...');
      const goVersion = '1.21.5';
      const arch = process.arch === 'arm64' ? 'arm64' : 'amd64';
      
      await Utils.runCommandWithOutput('wget', [
        `https://golang.org/dl/go${goVersion}.linux-${arch}.tar.gz`,
        '-O', 'go.tar.gz'
      ]);
      
      await Utils.runCommand('sudo', ['rm', '-rf', '/usr/local/go']);
      await Utils.runCommand('sudo', ['tar', '-C', '/usr/local', '-xzf', 'go.tar.gz']);
      await Utils.runCommand('rm', ['go.tar.gz']);

      // Add to PATH
      const bashrc = path.join(process.env.HOME, '.bashrc');
      const pathLine = 'export PATH=$PATH:/usr/local/go/bin';
      
      if (fs.existsSync(bashrc)) {
        const content = fs.readFileSync(bashrc, 'utf8');
        if (!content.includes('/usr/local/go/bin')) {
          fs.appendFileSync(bashrc, `\n${pathLine}\n`);
        }
      }
      
      process.env.PATH += ':/usr/local/go/bin';
    }

    // Install Node.js
    if (!Utils.commandExists('node')) {
      Utils.log.info('Installing Node.js...');
      await Utils.runCommand('curl', ['-fsSL', 
        'https://rpm.nodesource.com/setup_18.x', '|', 'sudo', 'bash', '-'
      ]);
      await Utils.runCommand('sudo', [pkgMgr, 'install', '-y', 'nodejs']);
    }

    // Install nmap
    if (!Utils.commandExists('nmap')) {
      Utils.log.info('Installing nmap...');
      await Utils.runCommand('sudo', [pkgMgr, 'install', '-y', 'nmap']);
    }
  }

  async installWindowsDeps() {
    Utils.log.warning('Windows installation requires manual setup:');
    console.log('1. Install Go: https://golang.org/dl/');
    console.log('2. Install Node.js: https://nodejs.org/');
    console.log('3. Install nmap: https://nmap.org/download.html');
    console.log('4. Ensure all tools are in your PATH');
    
    const { proceed } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'proceed',
        message: 'Have you installed Go, Node.js, and nmap?',
        default: false
      }
    ]);

    if (!proceed) {
      Utils.log.error('Please install the required dependencies and run the installer again');
      process.exit(1);
    }

    // Verify installations
    if (!Utils.commandExists('go')) throw new Error('Go not found in PATH');
    if (!Utils.commandExists('node')) throw new Error('Node.js not found in PATH');
    if (!Utils.commandExists('nmap')) throw new Error('nmap not found in PATH');
  }

  async setupRecoNya() {
    Utils.log.info('Setting up RecoNya...');

    // Go up one directory to get to project root (we're in scripts/ dir)
    const projectRoot = path.join(process.cwd(), '..');
    
    // Create .env file
    const envPath = path.join(projectRoot, 'backend', '.env');
    const envExamplePath = path.join(projectRoot, 'backend', '.env.example');

    if (fs.existsSync(envPath)) {
      Utils.log.success('.env file already exists');
    } else if (fs.existsSync(envExamplePath)) {
      Utils.log.info('Creating .env file from example...');
      fs.copyFileSync(envExamplePath, envPath);
      Utils.log.success('.env file created');
    } else {
      Utils.log.info('Creating default .env file...');
      const defaultEnv = `LOGIN_USERNAME=admin
LOGIN_PASSWORD=password
NETWORK_RANGE="192.168.1.0/24"
DATABASE_NAME="reconya-dev"
JWT_SECRET_KEY="${Utils.generateSecretKey()}"
SQLITE_PATH="data/reconya-dev.db"
`;
      fs.writeFileSync(envPath, defaultEnv);
      Utils.log.success('.env file created');
    }

    // Install Go dependencies
    Utils.log.info('Installing Go dependencies...');
    const backendPath = path.join(projectRoot, 'backend');
    await Utils.runCommand('go', ['mod', 'download'], { cwd: backendPath });

    // Install Node.js dependencies for frontend
    Utils.log.info('Installing frontend dependencies...');
    const frontendPath = path.join(projectRoot, 'frontend');
    await Utils.runCommand('npm', ['install'], { cwd: frontendPath });

    Utils.log.success('RecoNya setup complete!');
  }

  async createScripts() {
    Utils.log.info('Creating management scripts...');

    const projectRoot = path.join(process.cwd(), '..');
    const packageJsonPath = path.join(projectRoot, 'package.json');
    
    // Check if package.json already exists
    if (fs.existsSync(packageJsonPath)) {
      Utils.log.success('Package.json already exists');
      return;
    }

    const packageJson = {
      "name": "reconya",
      "version": "1.0.0",
      "description": "Network reconnaissance and asset discovery tool",
      "scripts": {
        "install": "cd scripts && npm install && node install.js",
        "start": "cd scripts && node start.js",
        "stop": "cd scripts && node stop.js",
        "status": "cd scripts && node status.js",
        "uninstall": "cd scripts && node uninstall.js"
      },
      "repository": {
        "type": "git",
        "url": "https://github.com/Dyneteq/reconya-ai-go.git"
      },
      "keywords": [
        "network",
        "reconnaissance",
        "security",
        "nmap",
        "asset-discovery"
      ],
      "author": "RecoNya",
      "license": "CC-BY-NC-4.0",
      "engines": {
        "node": ">=14.0.0"
      }
    };

    fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 2));
    Utils.log.success('Management scripts created');
  }
}

// Main execution
if (require.main === module) {
  program
    .name('reconya-install')
    .description('RecoNya installation script')
    .version('1.0.0');

  program.parse();

  const installer = new Installer();
  installer.run().catch(error => {
    Utils.log.error('Installation failed: ' + error.message);
    process.exit(1);
  });
}

module.exports = Installer;