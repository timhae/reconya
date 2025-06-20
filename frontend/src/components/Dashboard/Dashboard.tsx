// src/components/Dashboard/Dashboard.tsx
import React from 'react';
import useAuth from '../../hooks/useAuth';
import useDevices from '../../hooks/useDevices';
import useSystemStatus from '../../hooks/useSystemStatus';
import NetworkMap from '../NetworkMap/NetworkMap';
import Devices from '../Devices/Devices';
import DeviceList from '../DeviceList/DeviceList';
import EventLogs from '../EventLog/EventLogs';
import SystemStatusComponent from '../SystemStatus/SystemStatus';
import LoadingSpinner from '../Common/LoadingSpinner';
import useNetwork from '../../hooks/useNetwork';

const Dashboard: React.FC = () => {
  const isAuthenticated = useAuth();

  const { devices, isLoading: devicesLoading, error: devicesError, updateDeviceInState } = useDevices();
  const { systemStatus, isLoading: systemStatusLoading, error: systemStatusError } = useSystemStatus();
  const { network } = useNetwork();

  const localDevice = systemStatus?.local_device || systemStatus?.LocalDevice;

  const isLoading = devicesLoading || systemStatusLoading;
  const error = devicesError || systemStatusError;

  
  if (!isAuthenticated) return null;
  if (isLoading) return <LoadingSpinner message="Loading dashboard data..." />;
  if (error) return (
    <div className="alert alert-danger">
      <h4 className="alert-heading">Error</h4>
      <p>{error.message}</p>
      <hr />
      <p className="mb-0">Please try refreshing the page or contact support if the problem persists.</p>
    </div>
  );

  return (
    <div className="container-fluid">
      <div className="row mt-1">
        <div className="col-md-9">
          {/* Interactive D3.js Network Graph */}
          <NetworkMap devices={devices} localDevice={localDevice} />
          <Devices devices={devices} localDevice={localDevice} onDeviceUpdate={updateDeviceInState} />
        </div>
        <div className="col-md-3 d-flex flex-column">
          <div className="mb-4">
            <SystemStatusComponent systemStatus={systemStatus} network={network} />
          </div>
          <div className="flex-grow-1">
            <EventLogs />
          </div>
        </div>
      </div>
      <div className="mt-4">
        <DeviceList devices={devices} localDevice={localDevice} />
      </div>
    </div>
  );
};

export default Dashboard;
