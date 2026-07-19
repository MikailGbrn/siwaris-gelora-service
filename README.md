# Siwaris Gelora Service App

## 🚀 Overview

Siwaris Gelora Service is a comprehensive backend system built with **Laravel** that serves as the core engine for the Siwaris Gelora application. This service handles user management, real-time chat, event management, notifications, and various business logic operations through a robust RESTful API.

## ✨ Key Features

- **User Management**:
  - Authentication (Login, Register, Logout).
  - User profile management with image uploads.
  - Account status tracking (pending, active, banned).
- **Real-Time Chat**:
  - Real-time messaging between users.
  - Message history and status tracking (sent, delivered, read).
  - Online presence detection.
- **Event Management**:
  - Create and manage events with ticketing.
  - QR code generation for tickets.
  - Ticket validation and access control.
- **Notifications**:
  - In-app notifications for various events.
  - Push notifications (integration ready).
- **System Management**:
  - User management by administrators.
  - Role-based access control.

## 🔧 Technical Stack

- **Framework**: [Laravel](https://laravel.com/) (Latest Version)
- **Database**: MySQL
- **Authentication**: Laravel Sanctum (API Tokens)
- **Real-Time**: Laravel Echo (via Pusher)
- **Storage**: Amazon S3 (for file uploads)

## 📂 Project Structure

- `app/`:
  - `Http/Controllers`: API Controllers.
  - `Models`: Eloquent Models.
  - `Events`: Laravel Events for real-time broadcasting.
  - `Listeners`: Event Listeners.
- `database/migrations`: Database schema migrations.
- `routes/api.php`: API route definitions.

## ⚙️ Installation & Setup

### Prerequisites

- PHP 8.1 or higher
- MySQL
- Composer
- Node.js (for frontend/WebSocket testing)

### Installation Steps

1.  **Clone the repository**:
    ```bash
    git clone <repository-url>
    cd siwaris-gelora-service
    ```

2.  **Install dependencies**:
    ```bash
    composer install
    ```

3.  **Environment Configuration**:
    Copy the environment file and fill in your credentials:
    ```bash
    cp .env.example .env
    ```

    Edit `.env` with your database and service configurations:
    ```ini
    DB_CONNECTION=mysql
    DB_HOST=127.0.0.1
    DB_PORT=3306
    DB_DATABASE=siwaris_gelora
    DB_USERNAME=root
    DB_PASSWORD=
    ```

4.  **Generate Application Key**:
    ```bash
    php artisan key:generate
    ```

5.  **Run Migrations**:
    ```bash
    php artisan migrate
    ```

6.  **Start Development Server**:
    ```bash
    php artisan serve
    ```
    The API will be available at `http://localhost:8000`.

## 🚀 Usage

### Authentication

To access protected routes, you need an API token.

1.  **Login**:
    ```http
    POST /api/auth/login
    ```
    **Body**:
    ```json
    {
        "email": "[EMAIL_ADDRESS]",
        "password": "password"
    }
    ```

2.  **Get Token**:
    The response will contain an `access_token`. Use this in the `Authorization` header:
    ```http
    Authorization: Bearer <your-token>
    ```

### API Endpoints Overview

| Module | Routes | Description |
| :--- | :--- | :--- |
| **Auth** | `/api/auth/*` | Login, register, logout, profile. |
| **Users** | `/api/users/*` | User management and search. |
| **Chat** | `/api/chat/*` | Real-time messaging endpoints. |
| **Events** | `/api/events/*` | Event creation and ticket management. |
| **QR** | `/api/generate-qr` | QR code generation. |

## 🧪 Testing

### Running Tests

Run the test suite using PHPUnit:

```bash
php artisan test
```

## 🔌 WebSockets (Real-Time)

For real-time features like chat and notifications, ensure you have a WebSocket server running (e.g., Laravel Echo Server or Pusher).

**Start Broadcasting Server** (if using local server):
```bash
php artisan websockets:serve
```

## 🔐 Security

- Use HTTPS in production.
- Keep your `.env` file secure.
- Regularly rotate API tokens.

## 🤝 Contributing

1.  Fork the repository.
2.  Create a feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.

## 📄 License

This project is open-sourced software licensed under the [MIT license](https://opensource.org/licenses/MIT).

---

**Built with ❤️ using Laravel**
