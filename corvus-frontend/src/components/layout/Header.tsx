import { Link } from "react-router-dom";
import LogoPlaceholder from "./LogoPlaceholder";
import FriendCodeInput from "../shared/FriendCodeInput";

/** Header bar with logo and friend code input */
export default function Header() {
  return (
    <header className="border-b border-gray-200 bg-white">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 py-3 flex items-center justify-between">
        <Link to="/" className="no-underline text-black">
          <LogoPlaceholder />
        </Link>
        <FriendCodeInput />
      </div>
    </header>
  );
}

