import React from 'react';
import useDevices from '../../hooks/useDevices'; // Adjust the import path as needed
import NetworkMap from '../NetworkMap/NetworkMap'; // Adjust the import path as needed
import Devices from '../Devices/Devices'; // Adjust the import path as needed
import DeviceList from '../DeviceList/DeviceList';

const Dashboard: React.FC = () => {
  const { devices, isLoading, error } = useDevices();
  // Temporarily define localDevice or fetch it from somewhere
  const localDevice = undefined; // Placeholder, adjust as needed

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      <NetworkMap devices={devices} localDevice={localDevice} />
      <DeviceList devices={devices} localDevice={localDevice} />
      <Devices devices={devices} />
    </div>
  );
};

export default Dashboard;
