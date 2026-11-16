
import { Navigate } from "react-router-dom";
import { useProfileStore } from "../../stores/profileStore";
import { Loader } from "../../shared/components/UI/Loader";

export const AccountIndexRedirect = () => {
  const user = useProfileStore((state) => state.user);

  if (!user) {
    return <Loader />;
  }

  const redirectPath =
    user.role === "lead" ? "/account/intern" : "/account/chat";

  return <Navigate to={redirectPath} replace />;
};
