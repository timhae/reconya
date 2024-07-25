import React from 'react';
import { Device } from '../../models/device.model';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faSquare } from '@fortawesome/free-solid-svg-icons';

interface DeviceModalProps {
  device: Device | null;
  onClose: () => void;
}

const DeviceModal: React.FC<DeviceModalProps> = ({ device, onClose }) => {
  if (!device) return null;

  const calcTimeElapsed = (dateString: string | undefined) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: 'numeric', month: 'long', day: 'numeric',
    });
  };

  // Prevent modal close function from triggering when clicking inside the modal
  const handleModalContentClick = (e: React.MouseEvent) => {
    e.stopPropagation();
  };

  return (
    <div className="modal show d-block" tabIndex={-1} style={{ backgroundColor: 'rgba(0,0,0,0.5)', marginTop: '170px' }} onClick={onClose}>
      <div className="modal-dialog modal-lg" onClick={handleModalContentClick} style={{ minHeight: '400px' }}>
        <div className="modal-content">
          <div className="modal-body bg-black border border-success border-radius-0 text-success p-5" style={{ minHeight: '400px' }}>
          <div className="mb-3">
              <div className="border-bottom border-success pb-2 mb-3 d-flex justify-content-between">
                <div>
                  <span className="orbitron fw-bold fs-2">{device.IPv4}</span>
                </div>
                <div className="text-end">
                  <FontAwesomeIcon icon={faSquare} className="ms-3 fs-2 deviceFadeInAndOut" />
                </div>
              </div>
              <table className="text-success w-100 p-2 mb-4">
                <tbody className="p-2">
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Hostname</td>
                    <td>{device.Hostname || 'Unknown'}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">H/W vendor</td>
                    <td>{device.Vendor}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">MAC Address</td>
                    <td>{device.MAC}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Status</td>
                    <td>{device.Status}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">First appeared</td>
                    <td>{calcTimeElapsed(device.CreatedAt)}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Last seen online</td>
                    <td>{calcTimeElapsed(device.LastSeenOnlineAt)}</td>
                  </tr>
                </tbody>
              </table>

              <h6>[ PORTS ]</h6>
              <table className="text-success w-100 p-2 mb-4">
                <tbody className="p-2">
                  {device.Ports?.map((port, index) => (
                    <tr key={index}>
                      <td className="ps-2 fw-bold" style={{ width: '15%' }}>{port.number}</td>
                      <td style={{ width: '15%' }}>{port.state}</td>
                      <td style={{ width: '20%' }}>{port.protocol.toUpperCase()}</td>
                      <td>{port.service || 'Unknown'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {/* Event Log section can be added here based on your data and requirements */}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DeviceModal;
