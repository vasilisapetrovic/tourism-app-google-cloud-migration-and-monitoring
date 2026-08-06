import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import api from '../../api';

interface Execution {
  id: string;
  tourId: string;
  status: string;
  startedAt: string;
  endedAt?: string;
  completedKeypoints: { keypointId: string; name: string; completedAt: string }[];
}

interface Props {
  tourId?: string;
  tourStatus?: string;
  keypointsCount?: number;
  isPurchased?: boolean;
}

export default function TourExecution({ tourId: propTourId, tourStatus: propTourStatus, keypointsCount = 0, isPurchased = false }: Props) {
  const navigate = useNavigate();
  const { id: paramTourId } = useParams<{ id: string }>();
  const tourId = propTourId || paramTourId;
  const [execution, setExecution] = useState<Execution | null>(null);
  const [lastExecution, setLastExecution] = useState<Execution | null>(null);
  const [loading, setLoading] = useState(true);
  const [showStats, setShowStats] = useState(false);

  useEffect(() => {
    if (!tourId) return;
    const canStart = propTourStatus === 'published' || propTourStatus === 'archived';
    if (!canStart) { setLoading(false); return; }
    api.get('/executions/my')
      .then(res => {
        const all: Execution[] = Array.isArray(res.data) ? res.data : [];
        const active = all.find(e => e.tourId === tourId && e.status === 'active');
        const last = all
          .filter(e => e.tourId === tourId && (e.status === 'completed' || e.status === 'abandoned'))
          .sort((a, b) => new Date(b.startedAt).getTime() - new Date(a.startedAt).getTime())[0];
        setExecution(active || null);
        setLastExecution(last || null);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [tourId, propTourStatus]);

  const canStart = propTourStatus === 'published' || propTourStatus === 'archived';
  if (!canStart || loading) return null;

  const barStyle: React.CSSProperties = {
    position: 'fixed',
    bottom: 0,
    left: 250,
    right: 0,
    background: '#fff',
    borderTop: '1.5px solid #e5e7eb',
    padding: '18px 60px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 24,
    zIndex: 100,
    boxShadow: '0 -4px 24px rgba(0,0,0,0.07)',
  };

  const labelStyle: React.CSSProperties = {
    fontSize: 15,
    fontWeight: 700,
    color: '#000',
    display: 'block',
    marginBottom: 3,
    letterSpacing: '-0.02em',
  };

  const hintStyle: React.CSSProperties = {
    fontSize: 12,
    color: '#A5A5A3',
    margin: 0,
  };

  const btnPrimary: React.CSSProperties = {
    padding: '11px 24px',
    background: '#FA9819',
    color: '#fff',
    border: 'none',
    borderRadius: 6,
    fontSize: 13,
    fontWeight: 600,
    cursor: 'pointer',
    whiteSpace: 'nowrap',
    letterSpacing: '-0.01em',
  };

  const btnSecondary: React.CSSProperties = {
    padding: '11px 24px',
    background: 'transparent',
    color: '#000',
    border: '1.5px solid #e5e7eb',
    borderRadius: 6,
    fontSize: 13,
    fontWeight: 600,
    cursor: 'pointer',
    whiteSpace: 'nowrap',
    letterSpacing: '-0.01em',
  };

  if (execution) {
    return (
      <div style={barStyle}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div style={{ width: 8, height: 8, borderRadius: '50%', background: '#FA9819', boxShadow: '0 0 0 3px rgba(250,152,25,0.2)' }} />
          <div>
            <span style={labelStyle}>Tour in progress</span>
            <p style={hintStyle}>You have an active session for this tour.</p>
          </div>
        </div>
        <button style={btnPrimary} onClick={() => navigate(`/tours/${tourId}/execute`)}>
          Continue Tour
        </button>
      </div>
    );
  }

  if (lastExecution) {
    const isCompleted = lastExecution.status === 'completed';
    const startedAt = new Date(lastExecution.startedAt);
    const endedAt = lastExecution.endedAt ? new Date(lastExecution.endedAt) : new Date();
    const durationSec = Math.floor((endedAt.getTime() - startedAt.getTime()) / 1000);
    const durationStr = `${Math.floor(durationSec / 60)}m ${durationSec % 60}s`;
    const reached = lastExecution.completedKeypoints.length;

    if (showStats) {
      return (
        <div style={{ ...barStyle, flexDirection: 'column', alignItems: 'stretch', padding: '24px 60px', gap: 16 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ ...labelStyle, fontSize: 13, color: '#A5A5A3', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
              {isCompleted ? 'Completed' : 'Abandoned'} · Last session
            </span>
            <button style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#A5A5A3', fontSize: 18, padding: 0 }} onClick={() => setShowStats(false)}>✕</button>
          </div>
          <div style={{ display: 'flex', gap: 24, alignItems: 'center', flexWrap: 'wrap' }}>
            <div style={{ display: 'flex', gap: 32 }}>
              {[
                { val: durationStr, label: 'Duration' },
                { val: `${reached}/${keypointsCount}`, label: 'Key Points' },
                { val: `${keypointsCount > 0 ? Math.round((reached / keypointsCount) * 100) : 0}%`, label: 'Progress' },
                { val: endedAt.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }), label: isCompleted ? 'Finished' : 'Stopped' },
              ].map(s => (
                <div key={s.label} style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                  <span style={{ fontSize: 20, fontWeight: 700, color: '#FA9819', letterSpacing: '-0.03em' }}>{s.val}</span>
                  <span style={{ fontSize: 11, color: '#A5A5A3', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{s.label}</span>
                </div>
              ))}
            </div>
            <div style={{ marginLeft: 'auto', display: 'flex', gap: 10 }}>
              <button style={btnSecondary} onClick={() => setShowStats(false)}>Close</button>
              <button style={btnPrimary} onClick={() => navigate(`/tours/${tourId}/execute`)}>Start Again</button>
            </div>
          </div>
          {lastExecution.completedKeypoints.length > 0 && (
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', paddingTop: 4 }}>
              {lastExecution.completedKeypoints.map((ck, i) => (
                <span key={ck.keypointId} style={{ background: '#f1f5f9', border: '1px solid #e5e7eb', borderRadius: 20, padding: '3px 10px', fontSize: 12, color: '#000' }}>
                  {i + 1}. {ck.name}
                </span>
              ))}
            </div>
          )}
        </div>
      );
    }

    return (
      <div style={barStyle}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div style={{ width: 8, height: 8, borderRadius: '50%', background: isCompleted ? '#22c55e' : '#A5A5A3' }} />
          <div>
            <span style={labelStyle}>{isCompleted ? 'Tour completed' : 'Tour abandoned'}</span>
            <p style={hintStyle}>{durationStr} · {reached}/{keypointsCount} key points</p>
          </div>
        </div>
        <div style={{ display: 'flex', gap: 10 }}>
          <button style={btnSecondary} onClick={() => setShowStats(true)}>Statistics</button>
          <button style={btnPrimary} onClick={() => navigate(`/tours/${tourId}/execute`)}>Start Again</button>
        </div>
      </div>
    );
  }

  if (!isPurchased) {
    return (
      <div style={barStyle}>
        <div>
          <span style={labelStyle}>Purchase required</span>
          <p style={hintStyle}>You need to buy this tour before you can start it.</p>
        </div>
        <button style={{ ...btnPrimary, background: '#6b7280' }} onClick={() => navigate('/tours/explore')}>
          Go to Store
        </button>
      </div>
    );
  }

  return (
    <div style={barStyle}>
      <div>
        <span style={labelStyle}>Ready to explore</span>
        <p style={hintStyle}>Set your location in the Simulator before starting.</p>
      </div>
      <button style={btnPrimary} onClick={() => navigate(`/tours/${tourId}/execute`)}>
        Start Tour
      </button>
    </div>
  );
}