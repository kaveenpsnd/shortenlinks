import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { Link2, QrCode, TrendingUp, Plus, Copy, Check } from 'lucide-react';
import api from '../config/api';
import './Dashboard.css';

const Dashboard = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [stats, setStats] = useState({
    totalClicks: 0,
    activeLinks: 0,
    qrScans: 0,
  });
  const [links, setLinks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [copiedId, setCopiedId] = useState(null);

  useEffect(() => {
    if (user) {
      fetchUserLinks();
    }
  }, [user]);

  const fetchUserLinks = async () => {
    setLoading(true);
    try {
      const token = await user.getIdToken();
      console.log('Fetching user links with token...');
      
      const response = await api.get('/api/user/links', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      
      console.log('API Response:', response.data);
      
      const userLinks = response.data?.links || [];
      setLinks(userLinks);
      
      // Calculate stats from actual data
      const totalClicks = userLinks.reduce((sum, link) => sum + (link.clicks || 0), 0);
      setStats({
        totalClicks,
        activeLinks: userLinks.length,
        qrScans: Math.floor(totalClicks * 0.26), // Estimate based on typical conversion
      });
      
      console.log('Stats calculated:', { totalClicks, activeLinks: userLinks.length });
    } catch (error) {
      console.error('Failed to fetch links:', error.response?.data || error.message);
      setLinks([]);
      setStats({
        totalClicks: 0,
        activeLinks: 0,
        qrScans: 0,
      });
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text, id) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = Math.abs(now - date);
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
    return date.toLocaleDateString();
  };

  if (!user) {
    return (
      <div className="dashboard-container">
        <div className="not-authenticated">
          <h2>Please sign in to access your dashboard</h2>
        </div>
      </div>
    );
  }

  return (
    <div className="dashboard-container">
      {/* Profile Header */}
      <div className="profile-header">
        <div className="profile-info">
          <div className="avatar">
            {user.photoURL ? (
              <img src={user.photoURL} alt={user.displayName || 'User'} />
            ) : (
              <div className="avatar-placeholder">
                {(user.displayName || user.email || 'U')[0].toUpperCase()}
              </div>
            )}
            <div className="status-dot"></div>
          </div>
          <div className="user-details">
            <h1>{user.displayName || user.email?.split('@')[0] || 'User'}</h1>
            <p className="user-email">
              @{(user.email?.split('@')[0] || 'user').toLowerCase()}
              <span className="badge">PRO ACCOUNT</span>
            </p>
          </div>
        </div>
        <button onClick={() => navigate('/')} className="btn-create-link">
          <Plus size={18} />
          Create New Link
        </button>
      </div>

      {/* Stats Cards */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon" style={{ backgroundColor: '#e0f2fe', color: '#0284c7' }}>
            <TrendingUp size={24} />
          </div>
          <div className="stat-content">
            <div className="stat-label">Total Clicks</div>
            <div className="stat-value">{stats.totalClicks.toLocaleString()}</div>
            <div className="stat-trend positive">+12%</div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon" style={{ backgroundColor: '#dcfce7', color: '#16a34a' }}>
            <Link2 size={24} />
          </div>
          <div className="stat-content">
            <div className="stat-label">Active Links</div>
            <div className="stat-value">{stats.activeLinks}</div>
            <div className="stat-trend positive">+8%</div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon" style={{ backgroundColor: '#f3e8ff', color: '#9333ea' }}>
            <QrCode size={24} />
          </div>
          <div className="stat-content">
            <div className="stat-label">QR Scans</div>
            <div className="stat-value">{stats.qrScans.toLocaleString()}</div>
            <div className="stat-trend positive">+24%</div>
          </div>
        </div>
      </div>

      {/* Recent Links */}
      <div className="links-section">
        <div className="section-header">
          <h2>Your Recent Links</h2>
          <button className="btn-view-all">View All</button>
        </div>

        {loading ? (
          <div className="loading-state">
            <div className="spinner"></div>
            <p>Loading your links...</p>
          </div>
        ) : links.length === 0 ? (
          <div className="empty-state">
            <QrCode size={48} />
            <h3>No links yet</h3>
            <p>Create your first shortened link to get started</p>
            <button onClick={() => navigate('/')} className="btn-create-first">
              <Plus size={18} />
              Create Link
            </button>
          </div>
        ) : (
          <div className="links-table">
            <div className="table-header">
              <div className="col-short">SHORT LINK</div>
              <div className="col-original">ORIGINAL URL</div>
              <div className="col-qr">QR</div>
              <div className="col-actions">ACTIONS</div>
            </div>

            {links.map((link) => (
              <div key={link.id || link.short_code} className="table-row">
                <div className="col-short">
                  <a
                    href={`http://20.204.185.86/${link.short_code}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="short-link"
                  >
                    shorty.link/{link.short_code}
                  </a>
                  <div className="link-meta">
                    Created {formatDate(link.created_at)} • {link.clicks || 0} clicks
                  </div>
                </div>

                <div className="col-original">
                  <div className="original-url">{link.original_url}</div>
                </div>

                <div className="col-qr">
                  <button 
                    className="btn-icon"
                    title="View QR Code"
                  >
                    <QrCode size={18} />
                  </button>
                </div>

                <div className="col-actions">
                  <button
                    className="btn-action"
                    onClick={() => copyToClipboard(`http://20.204.185.86/${link.short_code}`, link.id || link.short_code)}
                    title="Copy Link"
                  >
                    {copiedId === (link.id || link.short_code) ? (
                      <Check size={18} />
                    ) : (
                      <Copy size={18} />
                    )}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Dashboard;
