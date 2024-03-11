// App.tsx
import React from 'react';
import './App.scss';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Navigation from './components/Common/Navigation';
import Dashboard from './components/Dashboard/Dashboard';
// Import other components as needed

function App() {
  return (
    <Router>
      <Navigation />
      <div className="fluid-container mx-auto px-5 mt-5 pb-5">
        <Routes>
          <Route path="/" element={<Dashboard />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
