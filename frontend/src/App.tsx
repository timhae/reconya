import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext'; // Adjust the path accordingly
import Navigation from './components/Common/Navigation';
import Dashboard from './components/Dashboard/Dashboard';
import Targets from './components/Targets/Targets';
import Login from './components/Login/Login';

import 'bootstrap/dist/js/bootstrap.bundle.min';
// import './App.scss';

function App() {
  return (
    <AuthProvider>
      <Router>
        <Navigation />
        <div className="fluid-container mx-auto px-5 mt-5 pb-5">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/targets/*" element={<Targets />} />
            <Route path="/login" element={<Login />} />
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
