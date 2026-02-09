import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { 
  TrendingUp, 
  Users, 
  Clock, 
  QrCode, 
  Copy, 
  Check, 
  Edit2,
  ChevronLeft,
  ChevronDown,
  Download
} from 'lucide-react';
import api from '../config/api';
import './Analytics.css';

const Analytics = () => {
  const { code } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const [linkData, setLinkData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);
  const [timeRange, setTimeRange] = useState('Last 30 Days');

  useEffect(() => {
    if (user && code) {
      fetchLinkStats();
    }
  }, [user, code]);

  const fetchLinkStats = async () => {
    setLoading(true);
    try {
      const token = await user.getIdToken();
      console.log('Fetching stats for code:', code);
      
      const response = await api.get(`/api/links/${code}/stats`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      
      console.log('Stats response:', response.data);
      setLinkData(response.data);
    } catch (error) {
      console.error('Failed to fetch link stats:', error.response?.data || error.message);
      setLinkData(null);
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  };

  // Mock data for visualization (in real app, this would come from backend)
  const mockLocations = [
    { country: 'United States', flag: '🇺🇸', count: 6420, percentage: 43 },
    { country: 'United Kingdom', flag: '🇬🇧', count: 2140, percentage: 17 },
    { country: 'Germany', flag: '🇩🇪', count: 1480, percentage: 12 },
    { country: 'Japan', flag: '🇯🇵', count: 890, percentage: 7 }
  ];

  const mockTraffic = [
    { source: 'Direct', percentage: 66, color: '#135bec' },
    { source: 'Social', percentage: 25, color: '#9333ea' },
    { source: 'Referral', percentage: 20, color: '#16a34a' }
  ];

  if (!user) {
    return (
      <div className="analytics-container">
        <div className="not-authenticated">
          <h2>Please sign in to view analytics</h2>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="analytics-container">
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Loading analytics...</p>
        </div>
      </div>
    );
  }

  if (!linkData) {
    return (
      <div className="analytics-container">
        <div className="not-found">
          <h2>Link not found</h2>
          <button onClick={() => navigate('/dashboard')}>Back to Dashboard</button>
        </div>
      </div>
    );
  }

  const shortUrl = `http://20.204.185.86/${linkData.short_code}`;
  const avgTimeOnPage = '1m 42s'; // Mock data
  const qrScans = Math.floor((linkData.clicks || 0) * 0.26);
  const uniqueVisitors = Math.floor((linkData.clicks || 0) * 0.71);

  return (
    <div className="analytics-container">
      {/* Breadcrumb */}
      <div className="breadcrumb">
        <Link to="/dashboard" className="breadcrumb-link">
          <ChevronLeft size={16} />
          Dashboard
        </Link>
        <span className="breadcrumb-separator">›</span>
        <span className="breadcrumb-current">Analytics</span>
      </div>

      {/* Link Header */}
      <div className="link-header">
        <div className="link-icon">
          <div className="icon-wrapper">
            📊
          </div>
        </div>
        <div className="link-info">
          <h1 className="link-title">shorty.link/{linkData.short_code}</h1>
          <p className="link-subtitle">
            Redirects to: {linkData.original_url}
          </p>
        </div>
        <div className="link-actions">
          <button 
            className="btn-secondary"
            onClick={() => copyToClipboard(shortUrl)}
          >
            {copied ? <Check size={18} /> : <Copy size={18} />}
            {copied ? 'Copied' : 'Copy'}
          </button>
          <button className="btn-primary">
            <Edit2 size={18} />
            Edit Link
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="analytics-stats-grid">
        <div className="analytics-stat-card">
          <div className="stat-header">
            <span className="stat-label">TOTAL CLICKS</span>
            <span className="stat-change positive">+12%</span>
          </div>
          <div className="stat-value-large">{(linkData.clicks || 0).toLocaleString()}</div>
        </div>

        <div className="analytics-stat-card">
          <div className="stat-header">
            <span className="stat-label">UNIQUE VISITORS</span>
            <span className="stat-change positive">+8%</span>
          </div>
          <div className="stat-value-large">{uniqueVisitors.toLocaleString()}</div>
        </div>

        <div className="analytics-stat-card">
          <div className="stat-header">
            <span className="stat-label">AVG. TIME ON PAGE</span>
            <span className="stat-change negative">-4%</span>
          </div>
          <div className="stat-value-large">{avgTimeOnPage}</div>
        </div>

        <div className="analytics-stat-card">
          <div className="stat-header">
            <span className="stat-label">QR SCANS</span>
            <span className="stat-change positive">+24%</span>
          </div>
          <div className="stat-value-large">{qrScans.toLocaleString()}</div>
        </div>
      </div>

      {/* Chart Section */}
      <div className="chart-section">
        <div className="section-header-inline">
          <h2>Clicks over Time</h2>
          <div className="time-range-selector">
            <button className="btn-time-range">
              {timeRange}
              <ChevronDown size={16} />
            </button>
          </div>
        </div>
        <p className="section-subtitle">Performance data for the last 30 days</p>
        
        <div className="chart-container">
          <svg className="line-chart" viewBox="0 0 800 200" preserveAspectRatio="none">
            <defs>
              <linearGradient id="areaGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                <stop offset="0%" stopColor="#135bec" stopOpacity="0.2" />
                <stop offset="100%" stopColor="#135bec" stopOpacity="0.05" />
              </linearGradient>
            </defs>
            
            {/* Area fill */}
            <path
              d="M 0 180 L 0 120 Q 100 140, 200 110 Q 300 80, 400 50 Q 500 30, 600 70 Q 700 90, 800 40 L 800 180 Z"
              fill="url(#areaGradient)"
            />
            
            {/* Line */}
            <path
              d="M 0 120 Q 100 140, 200 110 Q 300 80, 400 50 Q 500 30, 600 70 Q 700 90, 800 40"
              fill="none"
              stroke="#135bec"
              strokeWidth="3"
            />
            
            {/* Data points */}
            <circle cx="0" cy="120" r="4" fill="#135bec" />
            <circle cx="200" cy="110" r="4" fill="#135bec" />
            <circle cx="400" cy="50" r="4" fill="#135bec" />
            <circle cx="600" cy="70" r="4" fill="#135bec" />
            <circle cx="800" cy="40" r="4" fill="#135bec" />
          </svg>
          
          <div className="chart-x-axis">
            <span>OCT 25</span>
            <span>OCT 28</span>
            <span>OCT 31</span>
            <span>NOV 03</span>
            <span>NOV 06</span>
          </div>
        </div>
      </div>

      {/* Bottom Grid - Locations & Traffic */}
      <div className="analytics-bottom-grid">
        {/* Top Locations */}
        <div className="analytics-card">
          <div className="card-header-with-action">
            <h3>Top Locations</h3>
            <button className="btn-export">
              <Download size={16} />
              Export Data
            </button>
          </div>
          
          <div className="locations-list">
            {mockLocations.map((location, index) => (
              <div key={index} className="location-item">
                <div className="location-flag">{location.flag}</div>
                <div className="location-info">
                  <div className="location-name">{location.country}</div>
                  <div className="location-bar-container">
                    <div 
                      className="location-bar" 
                      style={{ width: `${location.percentage}%` }}
                    />
                  </div>
                </div>
                <div className="location-stats">
                  <div className="location-count">{location.count.toLocaleString()}</div>
                  <div className="location-percentage">({location.percentage}%)</div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Traffic Sources */}
        <div className="analytics-card">
          <h3>Traffic Sources</h3>
          
          <div className="donut-chart-container">
            <svg className="donut-chart" viewBox="0 0 200 200">
              <circle
                cx="100"
                cy="100"
                r="70"
                fill="none"
                stroke="#e7ebf3"
                strokeWidth="30"
              />
              <circle
                cx="100"
                cy="100"
                r="70"
                fill="none"
                stroke="#135bec"
                strokeWidth="30"
                strokeDasharray="293 440"
                strokeDashoffset="0"
                transform="rotate(-90 100 100)"
              />
              <circle
                cx="100"
                cy="100"
                r="70"
                fill="none"
                stroke="#9333ea"
                strokeWidth="30"
                strokeDasharray="110 440"
                strokeDashoffset="-293"
                transform="rotate(-90 100 100)"
              />
              <circle
                cx="100"
                cy="100"
                r="70"
                fill="none"
                stroke="#16a34a"
                strokeWidth="30"
                strokeDasharray="88 440"
                strokeDashoffset="-403"
                transform="rotate(-90 100 100)"
              />
              <text x="100" y="95" textAnchor="middle" className="donut-center-text">100%</text>
            </svg>
          </div>

          <div className="traffic-legend">
            {mockTraffic.map((source, index) => (
              <div key={index} className="traffic-item">
                <div className="traffic-dot" style={{ backgroundColor: source.color }} />
                <span className="traffic-label">{source.source}</span>
                <span className="traffic-percentage">{source.percentage}%</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Analytics;
