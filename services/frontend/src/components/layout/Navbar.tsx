import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';

export function Navbar() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <nav className="bg-white border-b border-slate-200 sticky top-0 z-50">
      <div className="max-w-6xl mx-auto px-4 h-14 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-lg font-bold text-blue-600">Pulse</span>
          <span className="text-lg font-light text-slate-500">Expends</span>
        </div>

        <div className="flex items-center gap-4">
          <span className="text-sm text-slate-600">
            {user?.display_name || user?.first_name || user?.email}
          </span>
          <button
            onClick={handleLogout}
            className="rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-100 transition"
          >
            Sign Out
          </button>
        </div>
      </div>
    </nav>
  );
}
