// components/Targets/AddTarget.tsx
import React from 'react';

const AddTarget: React.FC = () => {
  return (
    <div className="container mt-5">
      <ul className="nav nav-tabs" id="myTab" role="tablist">
        <li className="nav-item" role="presentation">
          <button className="nav-link active" id="domain-tab" data-bs-toggle="tab" data-bs-target="#domain" type="button" role="tab" aria-controls="domain" aria-selected="true">Add Domain</button>
        </li>
        <li className="nav-item" role="presentation">
          <button className="nav-link" id="ipAddress-tab" data-bs-toggle="tab" data-bs-target="#ipAddress" type="button" role="tab" aria-controls="ipAddress" aria-selected="false">Add IP Address</button>
        </li>
        <li className="nav-item" role="presentation">
          <button className="nav-link" id="multipleTargets-tab" data-bs-toggle="tab" data-bs-target="#multipleTargets" type="button" role="tab" aria-controls="multipleTargets" aria-selected="false">Add Multiple Targets</button>
        </li>
      </ul>
      <div className="tab-content" id="myTabContent">
        <div className="tab-pane fade show active" id="domain" role="tabpanel" aria-labelledby="domain-tab">
          <form className="mt-3">
            <div className="form-group">
              <label htmlFor="domainName">Domain name</label>
              <input type="text" className="form-control" id="domainName" placeholder="Enter domain name" />
            </div>
            <button type="submit" className="btn btn-primary mt-2">Submit</button>
          </form>
        </div>
        <div className="tab-pane fade" id="ipAddress" role="tabpanel" aria-labelledby="ipAddress-tab">
          <form className="mt-3">
            <div className="form-group">
              <label htmlFor="singleIp">IP Address</label>
              <input type="text" className="form-control" id="ipAddress" placeholder="Enter single IP address, CIDR or range" />
            </div>
            <button type="submit" className="btn btn-primary mt-2">Submit</button>
          </form>
        </div>
        <div className="tab-pane fade" id="multipleTargets" role="tabpanel" aria-labelledby="multipleTargets-tab">
          <form className="mt-3">
            <div className="form-group">
              <label htmlFor="multipleTargets">Multiple Targets</label>
              <textarea className="form-control" id="multipleTargets" placeholder="Enter multiple targets" />
            </div>
            <button type="submit" className="btn btn-primary mt-2">Submit</button>
          </form>
        </div>
      </div>
    </div>
  );
};

export default AddTarget;
