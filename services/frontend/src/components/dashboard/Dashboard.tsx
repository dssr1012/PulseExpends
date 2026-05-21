import { useState, useEffect } from 'react';
import { listCircles, createCircle } from '../../api/circles';
import type { Circle } from '../../types';

type ViewMode = 'personal' | 'family';

export function Dashboard() {
  const [mode, setMode] = useState<ViewMode>('personal');
  const [circles, setCircles] = useState<Circle[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadCircles();
  }, []);

  const loadCircles = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await listCircles();
      setCircles(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load circles');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateCircle = async () => {
    const name = mode === 'personal' ? 'My Finances' : 'Family Group';
    try {
      await createCircle({
        name,
        description: mode === 'personal' ? 'Personal expense tracking' : 'Family expense tracking',
        type: mode,
      });
      await loadCircles();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create circle');
    }
  };

  const filteredCircles = circles.filter((c) => c.type === mode);

  return (
    <div className="space-y-6">
      {/* Mode Toggle */}
      <div className="flex items-center gap-1 rounded-xl bg-slate-100 p-1 w-fit">
        <button
          onClick={() => setMode('personal')}
          className={`rounded-lg px-5 py-2 text-sm font-medium transition ${
            mode === 'personal'
              ? 'bg-white text-blue-600 shadow-sm'
              : 'text-slate-500 hover:text-slate-700'
          }`}
        >
          👤 Personal
        </button>
        <button
          onClick={() => setMode('family')}
          className={`rounded-lg px-5 py-2 text-sm font-medium transition ${
            mode === 'family'
              ? 'bg-white text-violet-600 shadow-sm'
              : 'text-slate-500 hover:text-slate-700'
          }`}
        >
          👨‍👩‍👧‍👦 Family
        </button>
      </div>

      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">
          {mode === 'personal' ? 'Personal Finances' : 'Family Finances'}
        </h1>
        <button
          onClick={handleCreateCircle}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 transition"
        >
          + New {mode === 'personal' ? 'Circle' : 'Group'}
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Circles List */}
      {isLoading ? (
        <div className="text-center py-12 text-slate-400">Loading…</div>
      ) : filteredCircles.length === 0 ? (
        <div className="text-center py-16">
          <div className="text-4xl mb-3">{mode === 'personal' ? '💰' : '👨‍👩‍👧‍👦'}</div>
          <h2 className="text-lg font-medium text-slate-700 mb-1">
            No {mode} circles yet
          </h2>
          <p className="text-sm text-slate-400 mb-4">
            Create your first {mode} circle to start tracking expenses
          </p>
          <button
            onClick={handleCreateCircle}
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 transition"
          >
            Create {mode === 'personal' ? 'Personal' : 'Family'} Circle
          </button>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filteredCircles.map((circle) => (
            <div
              key={circle.id}
              className="rounded-xl border border-slate-200 bg-white p-5 hover:shadow-md transition cursor-pointer"
            >
              <div className="flex items-start justify-between mb-3">
                <h3 className="font-semibold text-slate-900">{circle.name}</h3>
                <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                  circle.role === 'owner'
                    ? 'bg-blue-100 text-blue-700'
                    : circle.role === 'admin'
                    ? 'bg-violet-100 text-violet-700'
                    : 'bg-slate-100 text-slate-600'
                }`}>
                  {circle.role}
                </span>
              </div>
              <p className="text-sm text-slate-500 mb-4 line-clamp-2">{circle.description}</p>
              <div className="flex items-center gap-3 text-xs text-slate-400">
                <span>👥 {circle.member_count} member{circle.member_count !== 1 ? 's' : ''}</span>
                {circle.settings.monthly_budget && (
                  <span>🎯 ${circle.settings.monthly_budget.toLocaleString()} budget</span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
