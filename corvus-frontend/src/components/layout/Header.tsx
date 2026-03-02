import { Link } from "react-router-dom";
import LogoPlaceholder from "./LogoPlaceholder";
import FriendCodeInput from "../shared/FriendCodeInput";

export default function Header() {
  return (
    <header className="relative" style={{ zIndex: 20 }}>
      <div className="max-w-5xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
        <Link to="/" className="no-underline">
          <LogoPlaceholder />
        </Link>
        <FriendCodeInput />
      </div>
      {/* Brush stroke divider */}
      <div className="brush-divider-h" />
    </header>
  );
}
