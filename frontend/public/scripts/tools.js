export { apiRequest, timeAgo, PopupMessage}

// Reusable API request helper (JSON + credentials + error with status)
async function apiRequest(url, data, method = 'POST', extraHeaders = {}, log = false) {
  if (log) console.log({ url, data, method });

  try {
    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...extraHeaders,
      },
      credentials: 'include',
      body: JSON.stringify(data),
    });

    // If there's no content (204), return a special marker
    if (response.status === 204) return { status: 204, data: null };

    let responseData;
    const contentType = response.headers.get('Content-Type') || '';
    if (contentType.includes('application/json')) {
      responseData = await response.json();
    } else {
      responseData = await response.text();
    }

    return { status: response.status, data: responseData };

  } catch (err) {
    // Network or parsing error
    return { status: 0, error: err.message || 'Unknown error' };
  }
}





// add one listener to multiple event types on any EventTarget [element window document htmele ...]
EventTarget.prototype.addMultiEventListener = function (events, callback, options) {
  events.forEach(event => this.addEventListener(event, callback, options));
};



const timeAgo = (input) => {
  const date = input instanceof Date ? input : new Date(input)
  const formatter = new Intl.RelativeTimeFormat('en')
  const seconds = (Date.now() - date) / 1000

  const units = [
    ['year', 31536000],
    ['month', 2592000],
    ['week', 604800],
    ['day', 86400],
    ['hour', 3600],
    ['minute', 60],
    ['second', 1]
  ]

  for (const [unit, secondsInUnit] of units) {
    if (Math.abs(seconds) >= secondsInUnit) {
      return formatter.format(-Math.round(seconds / secondsInUnit), unit)
    }
  }

  return 'just now'
}

const PopupMessage = (message, state = 'fail', n = 10) => {
    let container = document.getElementById('popup-container');

    // Create container if it doesn't exist
    if (!container) {
        container = document.createElement('div');
        container.id = 'popup-container';
        document.body.appendChild(container);
    }

    const card = document.createElement('div');
    card.className = 'card ' + (state === 'success' ? 'success' : 'pperror');

    let iconPath = '';
    let titleText = '';

    if (state === 'success') {
        iconPath = 'M256 48a208 208 0 1 1 0 416 208 208 0 1 1 0-416zm0 464A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM369 209c9.4-9.4 9.4-24.6 0-33.9s-24.6-9.4-33.9 0l-111 111-47-47c-9.4-9.4-24.6-9.4-33.9 0s-9.4 24.6 0 33.9l64 64c9.4 9.4 24.6 9.4 33.9 0L369 209z';
        titleText = 'Success';
    } else if (state === 'info') {
        iconPath = 'M256 48a208 208 0 1 1 0 416 208 208 0 1 1 0-416zm0 344c13.3 0 24 10.7 24 24v8c0 13.3-10.7 24-24 24s-24-10.7-24-24v-8c0-13.3 10.7-24 24-24zm0-256c-17.7 0-32 14.3-32 32v96c0 17.7 14.3 32 32 32s32-14.3 32-32v-96c0-17.7-14.3-32-32-32z';
        titleText = 'Info';
    } else {
        iconPath = 'M256 48a208 208 0 1 1 0 416 208 208 0 1 1 0-416zm0 464A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM175 175c9.4-9.4 24.6-9.4 33.9 0L256 222.1 303 175c9.4-9.4 24.6-9.4 33.9 0s9.4 24.6 0 33.9L289.9 256 337 303c9.4 9.4 9.4 24.6 0 33.9s-24.6 9.4-33.9 0L256 289.9 209 337c-9.4 9.4-24.6 9.4-33.9 0s-9.4-24.6 0-33.9L222.1 256 175 209c-9.4-9.4-9.4-24.6 0-33.9z';
        titleText = 'Error';
    }

    card.innerHTML = `
        <div class="icon-container">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" fill="currentColor" class="icon">
                <path d="${iconPath}"/>
            </svg>
        </div>

        <div class="message-text-container">
            <p class="message-text">${titleText}</p>
            <p class="sub-text">${message}</p>
        </div>

        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 15 15" fill="none" class="cross-icon">
            <path fill="currentColor" d="M11.7816 4.03157C12.0062 3.80702 12.0062 3.44295 11.7816 3.2184C11.5571 2.99385 11.193 2.99385 10.9685 3.2184L7.50005 6.68682L4.03164 3.2184C3.80708 2.99385 3.44301 2.99385 3.21846 3.2184C2.99391 3.44295 2.99391 3.80702 3.21846 4.03157L6.68688 7.49999L3.21846 10.9684C2.99391 11.193 2.99391 11.557 3.21846 11.7816C3.44301 12.0061 3.80708 12.0061 4.03164 11.7816L7.50005 8.31316L10.9685 11.7816C11.193 12.0061 11.5571 12.0061 11.7816 11.7816C12.0062 11.557 12.0062 11.193 11.7816 10.9684L8.31322 7.49999L11.7816 4.03157Z"/>
        </svg>
    `;

    const closeBtn = card.querySelector('.cross-icon');
    closeBtn.addEventListener('click', () => {
        card.classList.add('removing');
        setTimeout(() => card.remove(), 300);
    });

    setTimeout(() => {
        if (card.parentNode) {
            card.classList.add('removing');
            setTimeout(() => card.remove(), 300*n);
        }
    }, 300*n);

    container.appendChild(card);
};




