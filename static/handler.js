function initCopyright() {
  document.querySelector('.copyright')
    .innerHTML = `COPYRIGHT © ${new Date().getFullYear().toString()} LSP. ALL RIGHTS RESERVED`;
}

function fetchConfig() {
  fetch('/get-config/')
    .then(resp => resp.json())
    .then(data => {
      dnsPanel.value = data['dns'];
      hostsPanel.value = data['hosts'];
    })
    .catch(() => {
      hostsPanel.value = dnsPanel.value = 'Failed to load config.';
    })
    .then(() => {
      [dnsPanel, hostsPanel].forEach((e, i) => {
        updateLineNum(e, lineNumPanel[i]);
        syncScroll(e, lineNumPanel[i], 1);
        e.addEventListener('keyup', () => updateLineNum(e, lineNumPanel[i]));
      });
    });
}

function changeBtnText(btn, text) {
  btn.innerHTML = text;
  setTimeout(() => btn.innerHTML = 'SAVE', 2000);
}

function validateDns(line) {
  let slices = line.trim().split('://');
  if (['tcp', 'udp', 'tcp-tls'].indexOf(slices[0]) !== -1) {
    slices = slices[1].split(':');
    if (!isNaN(slices[1]) && slices[1] > 0 && slices[1] < 65536) {
      if (ipRegex.test(slices[0])) {
        return true;
      }
    }
  }
}

function validateHost(line) {
  let slices = line.trim().split(/(\s+)/).filter(e => e.trim().length > 0);
  if (slices.length > 3) return false;
  if (ipRegex.test(slices[0])) {
    if (domainRegex.test(slices[1])) {
      return true;
    }
  }
}

function saveConfig(btn) {
  if (btn.innerHTML !== 'SAVE') return;

  const payload = {};
  if (btn === dnsSaveBtn) {
    if (lineNumPanel[0].value.indexOf('x') !== -1) {
      changeBtnText(btn, 'PLS CHECK ERROR(s)');
      return;
    }
    payload.dns = dnsPanel.value.trim();
  } else {
    if (lineNumPanel[1].value.indexOf('x') !== -1) {
      changeBtnText(btn, 'PLS CHECK ERROR(s)');
      return;
    }
    payload.hosts = hostsPanel.value.trim();
  }

  btn.innerHTML = 'SAVING';
  fetch('/save-config/', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  }).then((() => changeBtnText(btn, 'SUCCESS')))
    .catch(() => changeBtnText(btn, 'FAILED'));
}

function updateLineNum(contentPanel, numPanel) {
  numPanel.value = '';
  contentPanel.value.split('\n').forEach((line, idx) => {
    line = line.trim();
    let lineText = 'x';
    if (line === ''
      || line.startsWith('#')
      || (contentPanel === dnsPanel && validateDns(line))
      || (contentPanel === hostsPanel && validateHost(line))) {
      lineText = idx + 1;
    }
    numPanel.value += lineText + '\n';
  });
}

function syncScroll(l, r, scale) {
  let currentOver = l;
  l.addEventListener('scroll', () => {
    if (currentOver !== l) return;
    r.scrollTop = l.scrollTop * scale;
  });
  r.addEventListener('scroll', () => {
    if (currentOver !== r) return;
    l.scrollTop = r.scrollTop / scale;
  });
  l.addEventListener('mouseover', () => currentOver = l);
  r.addEventListener('mouseover', () => currentOver = r);
}

const dnsPanel = document.querySelector('.panel.dns .panel-setting');
const hostsPanel = document.querySelector('.panel.hosts .panel-setting');
const dnsSaveBtn = document.querySelector('.panel.dns .panel-submit');
const hostsSaveBtn = document.querySelector('.panel.hosts .panel-submit');
const lineNumPanel = document.querySelectorAll('.line-num');
const ipRegex = /^(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$/;
const domainRegex = /([a-z0-9]+\.)*[a-z0-9]+\.[a-z]+/;

dnsSaveBtn.addEventListener('click', () => saveConfig(dnsSaveBtn));
hostsSaveBtn.addEventListener('click', () => saveConfig(hostsSaveBtn));

initCopyright();
fetchConfig();