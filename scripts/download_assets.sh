#!/bin/bash
mkdir -p web/static/vendor

# Download React & ReactDOM
curl -L -o web/static/vendor/react.development.js https://unpkg.com/react@18/umd/react.development.js
curl -L -o web/static/vendor/react-dom.development.js https://unpkg.com/react-dom@18/umd/react-dom.development.js

# Download React Router
curl -L -o web/static/vendor/history.development.js https://unpkg.com/history@5/umd/history.development.js
curl -L -o web/static/vendor/react-router.development.js https://unpkg.com/react-router@6.3.0/umd/react-router.development.js
curl -L -o web/static/vendor/react-router-dom.development.js https://unpkg.com/react-router-dom@6.3.0/umd/react-router-dom.development.js

# Download Babel
curl -L -o web/static/vendor/babel.min.js https://unpkg.com/babel-standalone@6/babel.min.js

# Download Tailwind (Standalone CLI or Script version)
# Note: The script version is for dev only, but for this prototype it's fine.
curl -L -o web/static/vendor/tailwindcss.js "https://cdn.tailwindcss.com?plugins=forms,container-queries"

echo "Assets downloaded to web/static/vendor"
