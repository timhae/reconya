import React from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faGlobe, faNetworkWired } from '@fortawesome/free-solid-svg-icons';
import { SystemStatus } from '../../models/systemStatus.model';
import { Network } from '../../models/network.model';

const SystemStatusComponent: React.FC<{systemStatus: SystemStatus | undefined, network: Network | undefined}> = ({systemStatus, network}) => {
  // Use either the snake_case or camelCase property, depending on which is available
  const publicIP = systemStatus?.public_ip || systemStatus?.PublicIP;
  const networkCIDR = network?.cidr || network?.CIDR;
  
  console.log("SystemStatus component:", {systemStatus, network, publicIP, networkCIDR});
  
  return (
    <div>
      <div className="d-flex align-items-center mb-2">
        <h6 className="text-success d-block w-100">[ SYSTEM STATUS ]</h6>
      </div>
      <div className="card bg-very-dark text-success border-0 rounded-0">
        <div className="card-body">
          <div className="px-4 py-2 text-success fs-4">
            <div className="text-start mb-2">
              <div className="d-inline-block align-middle text-center" style={{width: 30}}>
                <FontAwesomeIcon icon={faNetworkWired} className="blink-online" />
              </div>
              <div className="d-inline-block align-middle ms-4">
                { networkCIDR }
              </div>
            </div>

            <div className="text-start">
              <div className="d-inline-block align-middle text-center" style={{width: 30}}>
                <FontAwesomeIcon icon={faGlobe} className="blink-active" />
              </div>
              <div className="d-inline-block align-middle ms-4">
                { publicIP }
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SystemStatusComponent;
