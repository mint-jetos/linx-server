document.getElementById('gatekeeper').addEventListener('statechange', (ev) => {
    if (ev.detail.state === 'verified') {
        fetch('/api/altcha/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ payload: ev.detail.payload })
        }).then(res => {
            if (res.ok) {
                window.location.reload();
            } else {
                alert('Verification failed. Please refresh.');
            }
        });
    }
});
