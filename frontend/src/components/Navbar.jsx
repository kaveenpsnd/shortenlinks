import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import AuthModal from './AuthModal';
import './Navbar.css';

const Navbar = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [authMode, setAuthMode] = useState('login'); // 'login' or 'signup'

  const handleOpenModal = (mode) => {
    setAuthMode(mode);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
  };

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  return (
    <>
      <header className="navbar">
        <div className="navbar-container">
          <Link to="/" className="navbar-brand">
            <img src="/logo.png" alt="Shrtner" className="navbar-logo" />
          </Link>
          
          <div className="navbar-actions">
            <nav className="navbar-nav">
              <Link to="/" className="nav-link">Home</Link>
              {user && (
                <Link to="/dashboard" className="nav-link">Dashboard</Link>
              )}
            </nav>
            
            <div className="navbar-buttons">
              {user ? (
                <button onClick={handleLogout} className="btn-logout">Logout</button>
              ) : (
                <>
                  <button onClick={() => handleOpenModal('login')} className="btn-login">
                    Login
                  </button>
                  <button onClick={() => handleOpenModal('signup')} className="btn-signup">
                    Sign Up
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      </header>

      {isModalOpen && (
        <AuthModal 
          isOpen={isModalOpen} 
          onClose={handleCloseModal} 
          initialMode={authMode}
        />
      )}
    </>
  );
};

export default Navbar;
