(function() {
    const gatekeeper = document.getElementById('gatekeeper');
    if (!gatekeeper) return;

    let verified = false;

    function handleVerified(payload) {
        if (verified) return;
        verified = true;
        console.log('ALTCHA verified, sending payload to server...');
        
        fetch('/api/altcha/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ payload: payload })
        }).then(res => {
            if (res.ok) {
                console.log('ALTCHA server verification success, forcing page reload...');
                // Force a reload from the server, bypassing cache
                window.location.replace(window.location.href);
            } else {
                console.error('ALTCHA server verification failed');
                alert('Verification failed. Please refresh.');
            }
        }).catch(err => {
            console.error('ALTCHA fetch error:', err);
        });
    }

    gatekeeper.addEventListener('statechange', (ev) => {
        console.log('ALTCHA state changed:', ev.detail.state);
        if (ev.detail.state === 'verified') {
            handleVerified(ev.detail.payload);
        }
    });

    // Check if it already finished before the script ran
    if (gatekeeper.state === 'verified') {
        console.log('ALTCHA already verified on load');
        handleVerified(gatekeeper.payload);
    }
})();
