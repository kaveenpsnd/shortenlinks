import React, { useState } from 'react';
import { 
  Zap, 
  QrCode, 
  BarChart3, 
  Link as LinkIcon, 
  Link2,
  ArrowRight
} from 'lucide-react';
import ResultCard from '../components/ResultCard';
import api from '../config/api';
import './Home.css';

const Home = () => {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState('');

  const handleShorten = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await api.post('/api/shorten', {
        original_url: url,
      });
      
      setResult({
        ...response.data,
        original_url: url
      });
      setUrl('');
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to shorten URL');
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="home-main">
      {/* Hero Section */}
      <section className="hero-section">
        <div className="hero-container">
          <div className="hero-badge">
            <span className="badge-dot"></span>
            <span className="badge-text">New: Advanced Analytics Dashboard</span>
          </div>
          
          <h1 className="hero-title">
            Shorten Your Links, <span className="text-primary">Expand Your Reach</span>
          </h1>
          
          <p className="hero-description">
            Transform long, ugly links into short, memorable URLs. Track clicks, analyze data, and optimize your marketing campaigns with our powerful platform.
          </p>
        </div>
      </section>

      {/* Input Section */}
      <section className="input-section">
        <div className="input-card">
          <form className="shorten-form" onSubmit={handleShorten}>
            <div className="input-wrapper">
              <div className="input-icon">
                <LinkIcon size={20} />
              </div>
              <input
                type="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="Paste your long link here..."
                className="url-input"
                required
              />
            </div>
            <button type="submit" className="btn-shorten" disabled={loading}>
              <span>{loading ? 'Shortening...' : 'Shorten Now'}</span>
              <ArrowRight size={18} />
            </button>
          </form>
          
          {error && <div className="error-message">{error}</div>}
        </div>

        {/* Result Card placed here if exists */}
        {result && (
          <div className="result-container">
            <ResultCard result={result} />
          </div>
        )}
      </section>

      {/* Features Grid */}
      <section className="features-section">
        <div className="features-container">
          <div className="features-grid">
            {/* Feature 1 */}
            <div className="feature-card">
              <div className="feature-icon-wrapper">
                <Zap className="feature-icon" size={24} />
              </div>
              <h3 className="feature-title">Lightning Fast</h3>
              <p className="feature-description">
                Experience instant redirection speeds with our optimized infrastructure designed for minimal latency.
              </p>
            </div>

            {/* Feature 2 */}
            <div className="feature-card">
              <div className="feature-icon-wrapper">
                <BarChart3 className="feature-icon" size={24} />
              </div>
              <h3 className="feature-title">Detailed Analytics</h3>
              <p className="feature-description">
                Gain insights into your audience with comprehensive data on clicks, locations, and devices.
              </p>
            </div>

            {/* Feature 3 */}
            <div className="feature-card">
              <div className="feature-icon-wrapper">
                <QrCode className="feature-icon" size={24} />
              </div>
              <h3 className="feature-title">QR Code Generation</h3>
              <p className="feature-description">
                Automatically generate customizable QR codes for every shortened link to bridge online and offline.
              </p>
            </div>
          </div>
        </div>
      </section>
      
      <footer className="footer_home">
        <div className="footer-content">
          <div className="footer-bottom">
            <p>© 2026 shrtner.link. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </main>
  );
};

export default Home;
