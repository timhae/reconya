// components/Targets/TargetsList.tsx
import React from 'react';
import { NavLink } from 'react-router-dom';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faSearch, faEdit, faTrashAlt } from '@fortawesome/free-solid-svg-icons';

const TargetsList: React.FC = () => {
  return (
    <div className="container">
      <div className="d-flex justify-content-between align-items-center mt-3 mb-4">
        <h2>Targets</h2>
        <NavLink to="/targets/add" className="btn btn-primary">Add New Target</NavLink>
      </div>

      {/* Targets table */}
      <table className="table table-bordered">
        <thead>
          <tr>
            <th>Domain Name</th>
            <th>Description</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {/* Populate table with target data */}
          <tr>
            <td>example.com</td>
            <td>Sample description</td>
            <td>
              <button className="btn btn-primary btn-sm me-1">
                <FontAwesomeIcon icon={faSearch} />
              </button>
              <button className="btn btn-warning btn-sm me-1">
                <FontAwesomeIcon icon={faEdit} />
              </button>
              <button className="btn btn-danger btn-sm">
                <FontAwesomeIcon icon={faTrashAlt} />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
};

export default TargetsList;
