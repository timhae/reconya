import React from 'react';
import { Device } from '../../models/device.model';

interface Props {
  devices: Device[];
  localDevice?: Device;
}

// Function to generate all IP addresses in a /24 range
const generateIpRange = (baseIp: string): string[] => {
  const ipParts = baseIp.split('.').map(Number);
  const base = ipParts.slice(0, 3).join('.');
  const ips = [];

  for (let i = 1; i <= 254; i++) {
    ips.push(`${base}.${i}`);
  }

  return ips;
};

// Function to get the CSS classes for each IP address based on its status
const getDeviceContainerCssClasses = (ip: string, devices: Device[], localDevice?: Device): string => {
  const device = devices.find((device) => device.IPv4 === ip);

  if (!device) {
    return 'border border-dark opacity-75 text-muted'; // Style for missing devices
  }

  if (device.Status === 'online' && device.IPv4 !== localDevice?.IPv4) {
    return 'border border-success text-success';
  } else if (device.IPv4 === localDevice?.IPv4) {
    return 'border border-primary text-primary';
  } else if (device.Status === 'offline' || device.Status === 'unknown') {
    return 'border border-dark text-dark';
  } else if (device.Status === 'idle') {
    return 'border border-success text-success opacity-50';
  }

  return 'border border-success'; // Default for existing devices
};

const NetworkMap: React.FC<Props> = ({ devices, localDevice }) => {
  const baseIp = localDevice?.IPv4 || '192.168.1.0'; // Adjust the base IP as necessary
  const ipRange = generateIpRange(baseIp);

  return (
    <div className="device-container d-flex align-items-start flex-wrap">
      <h6 className="text-success d-block w-100">[ NETWORK MAP ]</h6>
      {ipRange.map((ip, index) => {
        const isExisting = devices.some(device => device.IPv4 === ip);
        return (
          <button
            key={index}
            className={`device-mini-box bg-very-dark d-block ps-2 pe-0 m-1 ${getDeviceContainerCssClasses(ip, devices, localDevice)}`}
            title={ip}
            style={{
              borderColor: isExisting ? '#198754' : '#6c757d', // Green border for existing devices
            }}
          >
          </button>
        );
      })}
    </div>
  );
};

export default NetworkMap;
