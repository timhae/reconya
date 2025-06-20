import React, { useState } from "react";
import { Device, WebService } from "../../models/device.model";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCircle, faExternalLinkAlt, faGlobe, faLock, faTimes, faEdit, faSave, faTimes as faCancel } from "@fortawesome/free-solid-svg-icons";
import { updateDevice } from "../../api/axiosConfig";

interface DeviceModalProps {
  device: Device | null;
  onClose: () => void;
  onDeviceUpdate?: (updatedDevice: Device) => void;
}

const DeviceModal: React.FC<DeviceModalProps> = ({ device, onClose, onDeviceUpdate }) => {
  const [selectedScreenshot, setSelectedScreenshot] = useState<{url: string, screenshot: string} | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [editingName, setEditingName] = useState('');
  const [editingComment, setEditingComment] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  
  if (!device) return null;
  
  // Helper functions to normalize property access
  const getDeviceIPv4 = (d: Device) => d.ipv4 || d.IPv4 || '';
  const getDeviceMAC = (d: Device) => d.mac || d.MAC;
  const getDeviceVendor = (d: Device) => d.vendor || d.Vendor;
  const getDeviceType = (d: Device) => d.device_type || d.DeviceType;
  const getDeviceOS = (d: Device) => d.os || d.OS;
  const getDeviceHostname = (d: Device) => d.hostname || d.Hostname;
  const getDeviceStatus = (d: Device) => d.status || d.Status;
  const getDevicePorts = (d: Device) => d.ports || d.Ports || [];
  const getDeviceWebServices = (d: Device) => d.web_services || d.WebServices || [];
  const getDeviceCreatedAt = (d: Device) => d.created_at || d.CreatedAt;
  const getDeviceLastSeen = (d: Device) => d.last_seen_online_at || d.LastSeenOnlineAt;
  const getDeviceName = (d: Device) => d.name || d.Name || '';
  const getDeviceComment = (d: Device) => d.comment || d.Comment || '';
  const getDeviceID = (d: Device) => d.id || d.ID || '';

  const handleEditClick = () => {
    setEditingName(getDeviceName(device));
    setEditingComment(getDeviceComment(device));
    setIsEditing(true);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditingName('');
    setEditingComment('');
  };

  const handleSaveEdit = async () => {
    const deviceId = getDeviceID(device);
    if (!deviceId) {
      console.error('Device ID not found');
      return;
    }

    setIsSaving(true);
    try {
      const updateData = {
        name: editingName || undefined,
        comment: editingComment || undefined
      };

      const updatedDevice = await updateDevice(deviceId, updateData);
      
      if (onDeviceUpdate) {
        onDeviceUpdate(updatedDevice);
      }
      
      setIsEditing(false);
    } catch (error) {
      console.error('Failed to update device:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const calcTimeElapsed = (dateString: string | undefined) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const getStatusIcon = (status: string | undefined) => {
    if (!status) return null;

    let colorClass = "";
    switch (status.toLowerCase()) {
      case "online":
        colorClass = "text-success";
        break;
      case "offline":
        colorClass = "text-danger";
        break;
      case "idle":
        colorClass = "text-warning";
        break;
      default:
        colorClass = "text-muted";
        break;
    }

    return <FontAwesomeIcon icon={faCircle} className={`${colorClass} me-2`} />;
  };

  const getPortStateIcon = (state: string) => {
    if (state === "open") {
      return (
        <FontAwesomeIcon
          icon={faCircle}
          className="text-danger me-2"
          style={{ fontSize: "0.6rem" }}
        />
      );
    } else if (state === "filtered") {
      return (
        <FontAwesomeIcon
          icon={faCircle}
          className="text-warning me-2"
          style={{ fontSize: "0.6rem" }}
        />
      );
    }
    return null;
  };

  const getPortLink = (portNumber: number, ipAddress: string) => {
    const httpPorts = [80, 8080, 8000];
    const httpsPorts = [443, 8443];

    if (httpPorts.includes(portNumber)) {
      return `http://${ipAddress}:${portNumber}`;
    } else if (httpsPorts.includes(portNumber)) {
      return `https://${ipAddress}:${portNumber}`;
    }
    return null;
  };

  const renderPortIcons = () => {
    const ports = getDevicePorts(device);
    
    if (ports?.some((port) => port.state === "open")) {
      return (
        <FontAwesomeIcon
          icon={faCircle}
          className="text-danger me-2"
          style={{ fontSize: "1.2rem" }}
        />
      );
    }

    if (ports?.some((port) => port.state === "filtered")) {
      return (
        <FontAwesomeIcon
          icon={faCircle}
          className="text-warning me-2"
          style={{ fontSize: "1.2rem" }}
        />
      );
    }

    return (
      <FontAwesomeIcon
        icon={faCircle}
        className="text-success me-2"
        style={{ fontSize: "1.2rem" }}
      />
    );
  };

  const getWebServiceIcon = (protocol: string) => {
    if (protocol.toLowerCase() === 'https') {
      return <FontAwesomeIcon icon={faLock} className="text-success me-2" />;
    }
    return <FontAwesomeIcon icon={faGlobe} className="text-success me-2" />;
  };

  const getStatusBadgeColor = (statusCode: number) => {
    if (statusCode >= 200 && statusCode < 300) {
      return "bg-success";
    } else if (statusCode >= 400 && statusCode < 500) {
      return "bg-warning";
    } else if (statusCode >= 500) {
      return "bg-danger";
    }
    return "bg-secondary";
  };

  const formatFileSize = (bytes: number | undefined) => {
    if (!bytes) return "N/A";
    const kb = bytes / 1024;
    if (kb < 1024) {
      return `${kb.toFixed(1)} KB`;
    }
    const mb = kb / 1024;
    return `${mb.toFixed(1)} MB`;
  };

  const handleModalContentClick = (e: React.MouseEvent) => {
    e.stopPropagation();
  };

  return (
    <>
      <div
        className="modal show d-block"
        tabIndex={-1}
        style={{ backgroundColor: "rgba(0,0,0,0.5)" }}
        onClick={onClose}
      >
        <div
          className="modal-dialog modal-lg"
          onClick={handleModalContentClick}
          style={{ maxHeight: "90vh", marginTop: "5vh", marginBottom: "5vh" }}
        >
          <div className="modal-content" style={{ maxHeight: "90vh", display: "flex", flexDirection: "column" }}>
            <div
              className="modal-body bg-black border border-success border-radius-0 text-success p-5"
              style={{ overflowY: "auto", flexGrow: 1 }}
            >
            <div className="mb-3">
              <div className="border-bottom border-success pb-2 mb-3 d-flex justify-content-between align-items-center">
                <div className="d-flex align-items-center">
                  <span className="orbitron fw-bold fs-2">{getDeviceIPv4(device)}</span>
                </div>
                <div className="d-flex align-items-center">
                  {renderPortIcons()}
                  {!isEditing ? (
                    <button
                      type="button"
                      className="btn btn-outline-success btn-sm ms-3"
                      onClick={handleEditClick}
                      title="Edit device"
                    >
                      <FontAwesomeIcon icon={faEdit} />
                    </button>
                  ) : (
                    <div className="d-flex gap-2 ms-3">
                      <button
                        type="button"
                        className="btn btn-success btn-sm"
                        onClick={handleSaveEdit}
                        disabled={isSaving}
                        title="Save changes"
                      >
                        <FontAwesomeIcon icon={faSave} />
                      </button>
                      <button
                        type="button"
                        className="btn btn-outline-secondary btn-sm"
                        onClick={handleCancelEdit}
                        disabled={isSaving}
                        title="Cancel editing"
                      >
                        <FontAwesomeIcon icon={faCancel} />
                      </button>
                    </div>
                  )}
                  <button
                    type="button"
                    className="btn-close btn-close-white ms-3"
                    onClick={onClose}
                    aria-label="Close"
                  ></button>
                </div>
              </div>

              <table className="text-success w-100 p-2 mb-4">
                <tbody className="p-2">
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Name</td>
                    <td>
                      {isEditing ? (
                        <input
                          type="text"
                          className="form-control form-control-sm bg-dark text-success border-success"
                          value={editingName}
                          onChange={(e) => setEditingName(e.target.value)}
                          placeholder="Device name"
                        />
                      ) : (
                        getDeviceName(device) || "Unknown"
                      )}
                    </td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Comment</td>
                    <td>
                      {isEditing ? (
                        <textarea
                          className="form-control form-control-sm bg-dark text-success border-success"
                          rows={3}
                          value={editingComment}
                          onChange={(e) => setEditingComment(e.target.value)}
                          placeholder="Add a comment about this device"
                        />
                      ) : (
                        getDeviceComment(device) || "No comment"
                      )}
                    </td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Hostname</td>
                    <td>{getDeviceHostname(device) || "Unknown"}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">H/W vendor</td>
                    <td>{getDeviceVendor(device) || "Unknown"}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Device Type</td>
                    <td>
                      <span className="badge bg-dark border border-success text-success">
                        {getDeviceType(device) ? getDeviceType(device)?.replace('_', ' ').toUpperCase() : "Unknown"}
                      </span>
                    </td>
                  </tr>
                  {getDeviceOS(device) && (
                    <tr>
                      <td className="w-25 ps-2 fw-bold">Operating System</td>
                      <td>
                        {getDeviceOS(device)?.name || "Unknown"}
                        {getDeviceOS(device)?.version && ` ${getDeviceOS(device)?.version}`}
                        {getDeviceOS(device)?.confidence && (
                          <small className="text-muted ms-2">
                            ({getDeviceOS(device)?.confidence}% confidence)
                          </small>
                        )}
                      </td>
                    </tr>
                  )}
                  <tr>
                    <td className="w-25 ps-2 fw-bold">MAC Address</td>
                    <td>{getDeviceMAC(device) || "Unknown"}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Status</td>
                    <td>
                      {getStatusIcon(getDeviceStatus(device))}
                      {getDeviceStatus(device)}
                    </td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">First appeared</td>
                    <td>{calcTimeElapsed(getDeviceCreatedAt(device))}</td>
                  </tr>
                  <tr>
                    <td className="w-25 ps-2 fw-bold">Last seen online</td>
                    <td>{calcTimeElapsed(getDeviceLastSeen(device))}</td>
                  </tr>
                </tbody>
              </table>

              <h6>[ PORTS ]</h6>
              <table className="text-success w-100 p-2 mb-4">
                <tbody className="p-2">
                  {getDevicePorts(device)?.map((port, index) => {
                    const portNumber = Number(port.number); // Ensure port number is treated as a number
                    const portLink = getPortLink(portNumber, getDeviceIPv4(device));
                    return (
                      <tr key={index}>
                        <td className="ps-2 fw-bold" style={{ width: "15%" }}>
                          {portNumber}
                        </td>
                        <td style={{ width: "15%" }}>
                          <span className="badge bg-black border border-dark text-success">
                            {getPortStateIcon(port.state)}
                            {port.state}
                          </span>
                        </td>
                        <td style={{ width: "20%" }}>
                          {port.protocol.toUpperCase()}
                        </td>
                        <td>{port.service || "Unknown"}</td>
                        <td>
                          {portLink && (
                            <a
                              href={portLink}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="ms-2 text-light"
                            >
                              <FontAwesomeIcon
                                icon={faExternalLinkAlt}
                                className="text-success"
                              />
                            </a>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>

              {/* Web Services Section */}
              {getDeviceWebServices(device)?.length > 0 && (
                <>
                  <h6>[ WEB SERVICES ]</h6>
                  <div className="web-services-container">
                    {getDeviceWebServices(device)?.map((webService, index) => (
                      <div key={index} className="web-service-card mb-3 p-3 border border-success rounded">
                        <div className="row">
                          <div className="col-md-8">
                            <div className="d-flex align-items-center mb-2">
                              <span className="me-2">
                                {getWebServiceIcon(webService.protocol)}
                              </span>
                              <a
                                href={webService.url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-success text-decoration-none fw-bold"
                              >
                                {webService.url}
                                <FontAwesomeIcon
                                  icon={faExternalLinkAlt}
                                  className="ms-2"
                                  style={{ fontSize: "0.8rem" }}
                                />
                              </a>
                            </div>
                            <div className="mb-2">
                              <strong className="text-success">{webService.title || "No title"}</strong>
                            </div>
                            <div className="d-flex gap-3 text-muted small">
                              <span>
                                <span 
                                  className={`badge ${getStatusBadgeColor(webService.status_code)} text-dark`}
                                >
                                  {webService.status_code}
                                </span>
                              </span>
                              {webService.server && (
                                <span>Server: {webService.server}</span>
                              )}
                              <span>Size: {formatFileSize(webService.size)}</span>
                            </div>
                          </div>
                          <div className="col-md-4">
                            {webService.screenshot && (
                              <div className="screenshot-container text-center">
                                <img
                                  src={`data:image/png;base64,${webService.screenshot}`}
                                  alt={`Screenshot of ${webService.url}`}
                                  className="img-thumbnail"
                                  style={{ 
                                    maxWidth: "150px", 
                                    maxHeight: "100px", 
                                    objectFit: "contain",
                                    cursor: "pointer"
                                  }}
                                  onClick={() => {
                                    setSelectedScreenshot({
                                      url: webService.url,
                                      screenshot: webService.screenshot!
                                    });
                                  }}
                                />
                                <div className="small text-muted mt-1">Click to enlarge</div>
                              </div>
                            )}
                            {!webService.screenshot && (
                              <div className="text-center text-muted small">
                                <div className="border border-secondary rounded p-3" style={{ height: "100px", display: "flex", alignItems: "center", justifyContent: "center" }}>
                                  No screenshot available
                                </div>
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </>
              )}

              {/* Event Log section can be added here based on your data and requirements */}
            </div>
          </div>
        </div>
      </div>
    </div>

      {/* Screenshot Modal */}
      {selectedScreenshot && (
        <div 
          className="modal show d-block" 
          style={{ backgroundColor: 'rgba(0,0,0,0.8)', zIndex: 2000 }}
          onClick={() => setSelectedScreenshot(null)}
        >
          <div className="modal-dialog modal-xl modal-dialog-centered">
            <div className="modal-content bg-black border-success">
              <div className="modal-header border-success">
                <h5 className="modal-title text-success">
                  Screenshot - {selectedScreenshot.url}
                </h5>
                <button
                  type="button"
                  className="btn-close btn-close-white"
                  onClick={() => setSelectedScreenshot(null)}
                >
                  <FontAwesomeIcon icon={faTimes} />
                </button>
              </div>
              <div className="modal-body text-center p-2">
                <img
                  src={`data:image/png;base64,${selectedScreenshot.screenshot}`}
                  alt={`Screenshot of ${selectedScreenshot.url}`}
                  className="img-fluid"
                  style={{ 
                    maxWidth: "100%", 
                    maxHeight: "80vh", 
                    objectFit: "contain"
                  }}
                />
              </div>
              <div className="modal-footer border-success">
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => setSelectedScreenshot(null)}
                >
                  Close
                </button>
                <a
                  href={selectedScreenshot.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn btn-success"
                >
                  Open Website <FontAwesomeIcon icon={faExternalLinkAlt} className="ms-1" />
                </a>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default DeviceModal;
