import { Outlet } from 'react-router-dom';
import Navbar from './Navbar';

export default function Layout() {
  return (
    <div className="min-h-screen bg-[#fafafa]">
      <Navbar />
      <main className="max-w-[470px] mx-auto px-0 pt-[60px] pb-10">
        <Outlet />
      </main>
    </div>
  );
}
