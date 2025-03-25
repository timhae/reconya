import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { AuthProvider } from './contexts/AuthContext';
import Navigation from './components/Common/Navigation';
import Dashboard from './components/Dashboard/Dashboard';
import Targets from './components/Targets/Targets';
import Login from './components/Login/Login';
import ErrorBoundary from './components/Common/ErrorBoundary';

import 'bootstrap/dist/js/bootstrap.bundle.min';
// import './App.scss';

// Get app name from environment variable with fallback
const APP_NAME = process.env.REACT_APP_NAME || 'RecoNya';
const APP_VERSION = process.env.REACT_APP_VERSION || '1.0.0';

function App() {
  return (
    <ErrorBoundary>
      <HelmetProvider>
        <AuthProvider>
          <Router>
            {/* Common meta tags for all pages */}
            <Helmet>
              <meta charSet="utf-8" />
              <title>{APP_NAME} - Network Reconnaissance Tool</title>
              <meta name="description" content="A network device discovery and monitoring tool" />
              <meta name="viewport" content="width=device-width, initial-scale=1" />
              <meta name="theme-color" content="#000000" />
              <meta name="application-name" content={APP_NAME} />
              <meta name="application-version" content={APP_VERSION} />
            </Helmet>
            
            <Navigation />
            <div className="fluid-container mx-auto px-5 mt-5 pb-5">
              <Routes>
                <Route path="/" element={
                  <ErrorBoundary>
                    <Dashboard />
                  </ErrorBoundary>
                } />
                <Route path="/targets/*" element={
                  <ErrorBoundary>
                    <Targets />
                  </ErrorBoundary>
                } />
                <Route path="/login" element={<Login />} />
                
                {/* Add a 404 route */}
                <Route path="*" element={
                  <>
                    <Helmet>
                      <title>404 - Page Not Found | {APP_NAME}</title>
                    </Helmet>
                    <div className="d-flex justify-content-center align-items-center flex-column text-success">
                      <h1 className="text-success mb-4">404</h1>
                      <p>Page not found</p>
                      <a href="/" className="btn btn-success mt-3">Return to Dashboard</a>
                    </div>
                  </>
                } />
              </Routes>
            </div>
          </Router>
        </AuthProvider>
      </HelmetProvider>
    </ErrorBoundary>
  );
}

export default App;
