import React from 'react';
import { NavLink } from 'react-router-dom';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCircleNodes, faBell, faGear } from '@fortawesome/free-solid-svg-icons';

const Navigation: React.FC = () => {
  const isUserLoggedIn = true; // Replace this with actual logic

  return (
    <nav className="navbar navbar-expand-lg navbar-dark bg-very-dark">
      <div className="container-fluid px-5">
        <div className="d-flex align-items-center">
          <NavLink to="/" className="navbar-brand">
            <div className="logo text-dark bg-success border border-dark d-inline-block px-3 py-2">
              <FontAwesomeIcon icon={faCircleNodes} /> reconYa A.I.
            </div>
          </NavLink>
        </div>
        <div className="d-flex align-items-center">
          <NavLink to="#" className="text-success ms-3">
            <FontAwesomeIcon icon={faBell} />
          </NavLink>
          <NavLink to="#" className="text-success ms-3 me-4">
            <FontAwesomeIcon icon={faGear} />
          </NavLink>
          {isUserLoggedIn && (
            <div className="dropdown">
              <button
                className="btn btn-success dropdown-toggle"
                type="button"
                id="dropdownMenuButton"
                data-bs-toggle="dropdown"
                aria-expanded="false"
              >
                Welcome, human
              </button>
              <ul className="dropdown-menu dropdown-menu-end" aria-labelledby="dropdownMenuButton">
                <li><NavLink to="/targets" className="dropdown-item">Targets</NavLink></li>
                <li><NavLink to="/scans" className="dropdown-item">Scans</NavLink></li>
                <li><NavLink to="/account" className="dropdown-item">Account</NavLink></li>
                <li><NavLink to="/logout" className="dropdown-item">Logout</NavLink></li>
              </ul>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navigation;
