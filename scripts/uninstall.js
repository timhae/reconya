#!/usr/bin/env node

const { program } = require('commander');
const inquirer = require('inquirer');
const Utils = require('./utils');
const fs = require('fs');
const path = require('path');

class Uninstaller {
  constructor() {
    this.os = Utils.getOS();
  }

  async run() {
    console.log('==========================================');
    console.log('       reconYa Uninstall Script          ');
    console.log('==========================================\n');

    const { confirmUninstall } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'confirmUninstall',
        message: 'Are you sure you want to uninstall reconYa?',
        default: false
      }
    ]);

    if (!confirmUninstall) {
      Utils.log.info('Uninstall cancelled');
      return;
    }

    try {
      // Stop running processes
      await this.stopProcesses();

      // Remove application files
      await this.removeApplicationFiles();

      // Remove nmap permissions
      await this.removeNmapPermissions();

      // Ask about dependencies
      await this.handleDependencies();

      console.log('\n==========================================');
      Utils.log.success('reconYa uninstall completed!');
      console.log('==========================================\n');
      
      Utils.log.info('To completely remove this directory, run:');
      console.log(`  cd .. && rm -rf ${path.basename(process.cwd())}\n`);
      
      if (!Utils.isWindows()) {
        Utils.log.warning('Note: If you manually installed Go, you may need to:');
        Utils.log.warning('  - Remove Go PATH entries from your shell profile');
        Utils.log.warning('  - Source your shell profile or restart your terminal\n');
      }

    } catch (error) {
      Utils.log.error('Uninstall failed: ' + error.message);
      process.exit(1);
    }
  }

  async stopProcesses() {
    Utils.log.info('Stopping any running reconYa processes...');

    // Stop services on standard ports
    await Utils.killProcessByPort(3008, 'backend');
    await Utils.killProcessByPort(3000, 'frontend');

    // Kill any remaining reconYa processes
    try {
      if (Utils.isWindows()) {
        // Windows-specific process cleanup
        const { stdout } = await Utils.runCommandWithOutput('tasklist', ['/FI', 'IMAGENAME eq go.exe']);
        // Parse and kill relevant processes
      } else {
        // Unix-like process cleanup
        await Utils.runCommand('pkill', ['-f', 'go run.*cmd'], { silent: true });
        await Utils.runCommand('pkill', ['-f', 'react-scripts start'], { silent: true });
      }
    } catch {
      // Ignore errors - processes might not exist
    }

    Utils.log.success('Stopped running processes');
  }

  async removeApplicationFiles() {
    Utils.log.info('Removing reconYa application files...');

    // Remove generated files
    const filesToRemove = [
      'start-reconya.sh',
      'package.json'
    ];

    for (const file of filesToRemove) {
      const filePath = path.join(process.cwd(), file);
      if (fs.existsSync(filePath)) {
        fs.unlinkSync(filePath);
        Utils.log.info(`Removed ${file}`);
      }
    }

    // Ask about database and data files
    if (fs.existsSync(path.join(process.cwd(), 'backend', 'data'))) {
      const { removeData } = await inquirer.prompt([
        {
          type: 'confirm',
          name: 'removeData',
          message: 'Remove database and data files?',
          default: true
        }
      ]);

      if (removeData) {
        const dataPath = path.join(process.cwd(), 'backend', 'data');
        fs.rmSync(dataPath, { recursive: true, force: true });
        Utils.log.info('Removed backend/data directory');
      }
    }

    // Ask about frontend dependencies
    if (fs.existsSync(path.join(process.cwd(), 'frontend', 'node_modules'))) {
      const { removeNodeModules } = await inquirer.prompt([
        {
          type: 'confirm',
          name: 'removeNodeModules',
          message: 'Remove frontend dependencies (node_modules)?',
          default: true
        }
      ]);

      if (removeNodeModules) {
        const nodeModulesPath = path.join(process.cwd(), 'frontend', 'node_modules');
        fs.rmSync(nodeModulesPath, { recursive: true, force: true });
        Utils.log.info('Removed frontend/node_modules');
      }
    }

    // Remove script dependencies
    if (fs.existsSync(path.join(process.cwd(), 'scripts', 'node_modules'))) {
      const scriptsNodeModulesPath = path.join(process.cwd(), 'scripts', 'node_modules');
      fs.rmSync(scriptsNodeModulesPath, { recursive: true, force: true });
      Utils.log.info('Removed scripts/node_modules');
    }

    // Remove lock files
    const lockFiles = [
      'frontend/package-lock.json',
      'scripts/package-lock.json'
    ];

    for (const lockFile of lockFiles) {
      const lockPath = path.join(process.cwd(), lockFile);
      if (fs.existsSync(lockPath)) {
        fs.unlinkSync(lockPath);
        Utils.log.info(`Removed ${lockFile}`);
      }
    }

    Utils.log.success('reconYa application files removed');
  }

  async removeNmapPermissions() {
    if (Utils.isWindows()) {
      Utils.log.info('Windows detected - no nmap permissions to remove');
      return;
    }

    if (Utils.commandExists('nmap')) {
      Utils.log.info('Removing nmap setuid permissions...');
      try {
        const nmapPath = require('child_process').execSync('which nmap', { encoding: 'utf8' }).trim();
        await Utils.runCommand('sudo', ['chmod', 'u-s', nmapPath], { silent: true });
        Utils.log.info('Removed nmap setuid permissions');
      } catch {
        Utils.log.warning('Failed to remove nmap permissions');
      }
    }
  }

  async handleDependencies() {
    console.log();
    Utils.log.warning('The following section will ask about removing system dependencies');
    Utils.log.warning('Only remove these if you\'re sure they\'re not needed by other applications\n');

    const { removeDeps } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'removeDeps',
        message: 'Do you want to remove system dependencies (Go, Node.js, nmap)?',
        default: false
      }
    ]);

    if (!removeDeps) return;

    switch (this.os) {
      case 'macos':
        await this.uninstallMacOSDeps();
        break;
      case 'debian':
        await this.uninstallDebianDeps();
        break;
      case 'redhat':
        await this.uninstallRedHatDeps();
        break;
      case 'windows':
        await this.uninstallWindowsDeps();
        break;
    }
  }

  async uninstallMacOSDeps() {
    Utils.log.info('Available dependencies to remove on macOS:');
    console.log('  - nmap');
    console.log('  - node (Node.js)');
    console.log('  - go\n');
    Utils.log.warning('Note: These may be used by other applications\n');

    const dependencies = [
      { name: 'nmap', package: 'nmap' },
      { name: 'Node.js', package: 'node' },
      { name: 'Go', package: 'go' }
    ];

    for (const dep of dependencies) {
      const { remove } = await inquirer.prompt([
        {
          type: 'confirm',
          name: 'remove',
          message: `Remove ${dep.name}?`,
          default: false
        }
      ]);

      if (remove) {
        try {
          await Utils.runCommand('brew', ['uninstall', dep.package]);
          Utils.log.success(`Removed ${dep.name}`);
        } catch {
          Utils.log.warning(`Failed to remove ${dep.name}`);
        }
      }
    }
  }

  async uninstallDebianDeps() {
    Utils.log.info('Available dependencies to remove on Debian-based systems:');
    console.log('  - nmap');
    console.log('  - nodejs');
    console.log('  - go (manually installed)\n');
    Utils.log.warning('Note: These may be used by other applications\n');

    const { removeNmap } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'removeNmap',
        message: 'Remove nmap?',
        default: false
      }
    ]);

    if (removeNmap) {
      try {
        await Utils.runCommand('sudo', ['apt-get', 'remove', '-y', 'nmap']);
        Utils.log.success('Removed nmap');
      } catch {
        Utils.log.warning('Failed to remove nmap');
      }
    }

    const { removeNode } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'removeNode',
        message: 'Remove Node.js?',
        default: false
      }
    ]);

    if (removeNode) {
      try {
        await Utils.runCommand('sudo', ['apt-get', 'remove', '-y', 'nodejs', 'npm']);
        Utils.log.success('Removed Node.js');
      } catch {
        Utils.log.warning('Failed to remove Node.js');
      }
    }

    const { removeGo } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'removeGo',
        message: 'Remove Go (manually installed)?',
        default: false
      }
    ]);

    if (removeGo) {
      try {
        await Utils.runCommand('sudo', ['rm', '-rf', '/usr/local/go']);
        
        // Remove from PATH in bashrc
        const bashrcPath = path.join(process.env.HOME, '.bashrc');
        if (fs.existsSync(bashrcPath)) {
          let content = fs.readFileSync(bashrcPath, 'utf8');
          content = content.replace(/export PATH=\$PATH:\/usr\/local\/go\/bin\n?/g, '');
          fs.writeFileSync(bashrcPath, content);
          Utils.log.info('Removed Go from PATH in ~/.bashrc');
        }
        
        Utils.log.success('Removed Go');
      } catch {
        Utils.log.warning('Failed to remove Go');
      }
    }
  }

  async uninstallRedHatDeps() {
    const pkgMgr = Utils.commandExists('dnf') ? 'dnf' : 'yum';
    
    Utils.log.info('Available dependencies to remove on Red Hat-based systems:');
    console.log('  - nmap');
    console.log('  - nodejs');
    console.log('  - go (manually installed)\n');
    Utils.log.warning('Note: These may be used by other applications\n');

    const dependencies = [
      { name: 'nmap', package: 'nmap' },
      { name: 'Node.js', package: 'nodejs' }
    ];

    for (const dep of dependencies) {
      const { remove } = await inquirer.prompt([
        {
          type: 'confirm',
          name: 'remove',
          message: `Remove ${dep.name}?`,
          default: false
        }
      ]);

      if (remove) {
        try {
          await Utils.runCommand('sudo', [pkgMgr, 'remove', '-y', dep.package]);
          Utils.log.success(`Removed ${dep.name}`);
        } catch {
          Utils.log.warning(`Failed to remove ${dep.name}`);
        }
      }
    }

    // Handle manually installed Go
    const { removeGo } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'removeGo',
        message: 'Remove Go (manually installed)?',
        default: false
      }
    ]);

    if (removeGo) {
      try {
        await Utils.runCommand('sudo', ['rm', '-rf', '/usr/local/go']);
        
        // Remove from PATH
        const bashrcPath = path.join(process.env.HOME, '.bashrc');
        if (fs.existsSync(bashrcPath)) {
          let content = fs.readFileSync(bashrcPath, 'utf8');
          content = content.replace(/export PATH=\$PATH:\/usr\/local\/go\/bin\n?/g, '');
          fs.writeFileSync(bashrcPath, content);
          Utils.log.info('Removed Go from PATH in ~/.bashrc');
        }
        
        Utils.log.success('Removed Go');
      } catch {
        Utils.log.warning('Failed to remove Go');
      }
    }
  }

  async uninstallWindowsDeps() {
    Utils.log.warning('Windows dependency removal must be done manually:');
    console.log('1. Go to "Add or Remove Programs" in Windows Settings');
    console.log('2. Search for and uninstall:');
    console.log('   - Go Programming Language');
    console.log('   - Node.js');
    console.log('   - Nmap');
    console.log('3. Remove any PATH entries manually if needed');
  }
}

// Main execution
if (require.main === module) {
  program
    .name('reconya-uninstall')
    .description('reconYa uninstall script')
    .version('1.0.0');

  program.parse();

  const uninstaller = new Uninstaller();
  uninstaller.run().catch(error => {
    Utils.log.error('Uninstall failed: ' + error.message);
    process.exit(1);
  });
}

module.exports = Uninstaller;