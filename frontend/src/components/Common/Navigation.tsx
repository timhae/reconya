// Navigation.tsx
import React from 'react';
import { NavLink } from 'react-router-dom';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCircleNodes, faBell, faGear } from '@fortawesome/free-solid-svg-icons';

const Navigation: React.FC = () => {
  // Assume `isUserLoggedIn` is determined by your authentication logic
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
            <div className="d-inline-block ml-auto">
              {/* Implement dropdown logic as needed */}
              <NavLink to="#" className="nav-link text-success">
                Welcome, human
              </NavLink>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navigation;
