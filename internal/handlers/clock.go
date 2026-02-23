package handlers

import (
	"net/http"
)

// ClockHandler serves the @clock.html content.
func ClockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Digital Clock</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;700&display=swap');

        body {
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            background: #1a1a1a; /* Luxury Dark Mode Background */
            font-family: 'Inter', sans-serif;
            overflow: hidden;
            margin: 0;
            color: #e0e0e0;
        }

        .clock-container {
            background: rgba(255, 255, 255, 0.05); /* Glassmorphism/Neumorphism base */
            border-radius: 16px;
            box-shadow: 0 4px 30px rgba(0, 0, 0, 0.1);
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            padding: 40px 60px;
            text-align: center;
            position: relative;
            overflow: hidden;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
        }

        .clock-container::before {
            content: '';
            position: absolute;
            top: -50%;
            left: -50%;
            width: 200%;
            height: 200%;
            background: radial-gradient(circle at top left, rgba(255, 255, 255, 0.1) 0%, transparent 70%);
            transform: rotate(45deg);
        }

        .time {
            font-size: 8em;
            font-weight: 700;
            letter-spacing: 5px;
            text-shadow: 0 0 10px rgba(0, 255, 255, 0.5), 0 0 20px rgba(0, 255, 255, 0.3); /* Subtle glow */
            transition: all 0.5s ease-in-out;
            margin-bottom: 20px;
        }

        .date {
            font-size: 2em;
            font-weight: 300;
            letter-spacing: 2px;
            color: rgba(255, 255, 255, 0.7);
            margin-top: -15px;
        }

        @media (max-width: 768px) {
            .time {
                font-size: 5em;
            }
            .date {
                font-size: 1.5em;
            }
            .clock-container {
                padding: 30px 40px;
            }
        }

        @media (max-width: 480px) {
            .time {
                font-size: 3em;
            }
            .date {
                font-size: 1em;
            }
            .clock-container {
                padding: 20px 25px;
            }
        }
    </style>
</head>
<body>
    <div class="clock-container">
        <div id="time" class="time"></div>
        <div id="date" class="date"></div>
    </div>

    <script>
        function updateClock() {
            const now = new Date();
            let hours = now.getHours();
            let minutes = now.getMinutes();
            let seconds = now.getSeconds();

            hours = hours < 10 ? '0' + hours : hours;
            minutes = minutes < 10 ? '0' + minutes : minutes;
            seconds = seconds < 10 ? '0' + seconds : seconds;

            document.getElementById('time').textContent = hours + ':' + minutes + ':' + seconds;

            const options = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };
            document.getElementById('date').textContent = now.toLocaleDateString(undefined, options);
        }

        setInterval(updateClock, 1000);
        updateClock(); // Initial call to display clock immediately
    </script>
</body>
</html>`))
}
