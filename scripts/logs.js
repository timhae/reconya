#!/usr/bin/env node

const { program } = require('commander');
const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');
const chalk = require('chalk');

class LogViewer {
  constructor() {
    this.logsDir = path.join(process.cwd(), 'logs');
    this.logFile = path.join(this.logsDir, 'reconya.log');
    this.errorFile = path.join(this.logsDir, 'reconya.error.log');
  }

  async viewLogs(options = {}) {
    const { follow, errors, lines } = options;
    
    console.log('==========================================');
    console.log('         reconYa Daemon Logs             ');
    console.log('==========================================\n');

    // Check if logs directory exists
    if (!fs.existsSync(this.logsDir)) {
      console.log(chalk.yellow('No logs directory found. Service may not be running in daemon mode.'));
      return;
    }

    const targetFile = errors ? this.errorFile : this.logFile;
    const logType = errors ? 'Error' : 'Application';

    if (!fs.existsSync(targetFile)) {
      console.log(chalk.yellow(`No ${logType.toLowerCase()} log file found: ${targetFile}`));
      return;
    }

    console.log(chalk.gray(`Viewing ${logType} logs: ${targetFile}\n`));

    if (follow) {
      // Follow mode - tail -f equivalent
      console.log(chalk.gray('Following logs (Press Ctrl+C to exit)...\n'));
      
      // Show last few lines first
      if (lines) {
        await this.showLastLines(targetFile, lines);
      }
      
      // Then follow
      this.followLogs(targetFile);
    } else {
      // Static view
      if (lines) {
        await this.showLastLines(targetFile, lines);
      } else {
        await this.showAllLogs(targetFile);
      }
    }
  }

  async showLastLines(filePath, numLines) {
    return new Promise((resolve, reject) => {
      const tail = spawn('tail', ['-n', numLines.toString(), filePath]);
      
      tail.stdout.on('data', (data) => {
        process.stdout.write(data.toString());
      });
      
      tail.stderr.on('data', (data) => {
        console.error(chalk.red(data.toString()));
      });
      
      tail.on('close', (code) => {
        if (code === 0) {
          resolve();
        } else {
          reject(new Error(`tail process exited with code ${code}`));
        }
      });
    });
  }

  async showAllLogs(filePath) {
    return new Promise((resolve, reject) => {
      const cat = spawn('cat', [filePath]);
      
      cat.stdout.on('data', (data) => {
        process.stdout.write(data.toString());
      });
      
      cat.stderr.on('data', (data) => {
        console.error(chalk.red(data.toString()));
      });
      
      cat.on('close', (code) => {
        if (code === 0) {
          resolve();
        } else {
          reject(new Error(`cat process exited with code ${code}`));
        }
      });
    });
  }

  followLogs(filePath) {
    // Create the file if it doesn't exist
    if (!fs.existsSync(filePath)) {
      fs.writeFileSync(filePath, '');
    }
    
    const tail = spawn('tail', ['-f', filePath], {
      stdio: ['ignore', 'pipe', 'pipe']
    });
    
    tail.stdout.on('data', (data) => {
      process.stdout.write(data.toString());
    });
    
    tail.stderr.on('data', (data) => {
      console.error(chalk.red(data.toString()));
    });
    
    tail.on('error', (error) => {
      console.error(chalk.red('tail process error: ' + error.message));
    });
    
    // Handle Ctrl+C gracefully
    const handleExit = () => {
      console.log(chalk.gray('\n\nStopping log viewer...'));
      tail.kill('SIGTERM');
      process.exit(0);
    };
    
    process.on('SIGINT', handleExit);
    process.on('SIGTERM', handleExit);
    
    tail.on('close', (code) => {
      if (code !== 0) {
        console.log(chalk.gray(`\nLog viewer exited with code ${code}`));
      }
      process.exit(code || 0);
    });
    
    // Keep the process alive
    process.stdin.resume();
  }

  async clearLogs() {
    console.log('Clearing daemon logs...');
    
    try {
      if (fs.existsSync(this.logFile)) {
        fs.writeFileSync(this.logFile, '');
        console.log(chalk.green('Application logs cleared'));
      }
      
      if (fs.existsSync(this.errorFile)) {
        fs.writeFileSync(this.errorFile, '');
        console.log(chalk.green('Error logs cleared'));
      }
      
      console.log(chalk.green('All logs cleared successfully'));
    } catch (error) {
      console.error(chalk.red('Failed to clear logs: ' + error.message));
      process.exit(1);
    }
  }
}

// Main execution
if (require.main === module) {
  program
    .name('reconya-logs')
    .description('View reconYa daemon logs')
    .version('1.0.0')
    .option('-f, --follow', 'follow log output (like tail -f)')
    .option('-e, --errors', 'show error logs instead of application logs')
    .option('-n, --lines <number>', 'number of lines to show', parseInt)
    .option('-c, --clear', 'clear all log files')
    .parse();

  const options = program.opts();
  const logViewer = new LogViewer();

  if (options.clear) {
    logViewer.clearLogs();
  } else {
    logViewer.viewLogs(options).catch(error => {
      console.error(chalk.red('Failed to view logs: ' + error.message));
      process.exit(1);
    });
  }
}

module.exports = LogViewer;