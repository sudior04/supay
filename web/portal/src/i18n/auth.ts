export const authMessages = {
  vi: {
    passwordChanged: "Đổi mật khẩu thành công. Vui lòng đăng nhập lại.",
    back: "← Trang giới thiệu", eyebrow: "Supay / Tài khoản", loginTitle: "Chào mừng trở lại", registerTitle: "Tạo tài khoản Supay",
    loginDescription: "Đăng nhập để tiếp tục mua sắm và quản lý cửa hàng của bạn.", registerDescription: "Đăng ký miễn phí để mua sắm và sử dụng ví Supay.",
    name: "Họ và tên", namePlaceholder: "Nguyễn Minh Anh", email: "Email", emailPlaceholder: "you@example.com", password: "Mật khẩu",
    passwordPlaceholder: "Nhập mật khẩu", passwordHint: "Tối thiểu 10 ký tự", working: "Đang xử lý…", login: "Đăng nhập", register: "Tạo tài khoản",
    noAccount: "Chưa có tài khoản?", hasAccount: "Đã có tài khoản?", registerLink: "Đăng ký", loginLink: "Đăng nhập", walletNote: "Sau khi đăng ký, bạn có thể mở ví Supay VND từ không gian tài khoản của mình.",
    invalidLogin: "Email hoặc mật khẩu chưa chính xác.", invalidRegister: "Chưa thể tạo tài khoản với thông tin này.", missingRole: "Tài khoản đã xác thực nhưng hệ thống chưa trả về vai trò. Vui lòng thử đăng nhập lại.", network: "Không kết nối được dịch vụ tài khoản. Vui lòng thử lại.", language: "Ngôn ngữ"
  },
  en: {
    passwordChanged: "Your password was changed. Please sign in again.",
    back: "← Home", eyebrow: "Supay / Account", loginTitle: "Welcome back", registerTitle: "Create your Supay account",
    loginDescription: "Sign in to continue shopping and manage your store.", registerDescription: "Create a free account to shop and use Supay Wallet.",
    name: "Full name", namePlaceholder: "Alex Morgan", email: "Email", emailPlaceholder: "you@example.com", password: "Password",
    passwordPlaceholder: "Enter your password", passwordHint: "At least 10 characters", working: "Working…", login: "Sign in", register: "Create account",
    noAccount: "New to Supay?", hasAccount: "Already have an account?", registerLink: "Create account", loginLink: "Sign in", walletNote: "After registering, open a Supay VND wallet from your account space.",
    invalidLogin: "That email and password combination was not recognized.", invalidRegister: "We could not create an account with those details.", missingRole: "Your account is verified, but no role was returned. Please try signing in again.", network: "Could not connect to the account service. Please try again.", language: "Language"
  }
} as const;
