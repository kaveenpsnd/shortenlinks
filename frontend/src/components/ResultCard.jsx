import React, { useState } from 'react';
import { Copy, Share2, Check } from 'lucide-react';
import { QRCodeCanvas } from 'qrcode.react';
import './ResultCard.css';

const ResultCard = ({ result }) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(result.short_url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: 'Shortened URL',
          text: 'Check out this shortened link!',
          url: result.short_url,
        });
      } catch (err) {
        console.log('Share failed:', err);
      }
    }
  };

  return (
    <div className="result-card">
      <div className="result-content">
        <div className="qr-code-container">
          <QRCodeCanvas
            value={result.short_url}
            size={120}
            level="M"
            includeMargin={true}
          />
        </div>

        <div className="result-details">
          <div className="result-label">SHORT LINK</div>
          <div className="result-url">
            <a href={result.short_url} target="_blank" rel="noopener noreferrer" className="short-url-link">
              {result.short_url} <Check className="check-icon" />
            </a>
          </div>
          
          <div className="original-url">
            <span className="url-label">Original:</span>
            <span className="url-text">{result.original_url || 'N/A'}</span>
          </div>

          <div className="result-actions">
            <button onClick={handleCopy} className="btn-copy">
              {copied ? <Check size={18} /> : <Copy size={18} />}
              <span>{copied ? 'Copied!' : 'Copy Link'}</span>
            </button>
            
            <button onClick={handleShare} className="btn-share">
              <Share2 size={18} />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ResultCard;
