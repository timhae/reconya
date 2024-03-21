import React from 'react';
import useDevices from '../../hooks/useDevices';
import useSystemStatus from '../../hooks/useSystemStatus';
import NetworkMap from '../NetworkMap/NetworkMap';
import Devices from '../Devices/Devices';
import DeviceList from '../DeviceList/DeviceList';

const Dashboard: React.FC = () => {
  const { devices, isLoading: devicesLoading, error: devicesError } = useDevices();
  // Destructure systemStatus and then access localDevice from it
  const { systemStatus, isLoading: systemStatusLoading, error: systemStatusError } = useSystemStatus();

  // Access localDevice directly from systemStatus; adjust the casing to match your actual data structure
  const localDevice = systemStatus?.LocalDevice; // Assuming localDevice is the correct property name

  const isLoading = devicesLoading || systemStatusLoading;
  const error = devicesError || systemStatusError;

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      <NetworkMap devices={devices} localDevice={localDevice} />
      <Devices devices={devices} localDevice={localDevice} />
      <DeviceList devices={devices} localDevice={localDevice} />
    </div>
  );
};

export default Dashboard;
