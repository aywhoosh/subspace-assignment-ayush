(function(){
  function qs(sel){return document.querySelector(sel)}
  function openModal(){
    var m = qs('#connect-modal');
    if (!m) return;
    m.setAttribute('aria-hidden','false');
    var t = m.querySelector('[data-testid="connect-note"]');
    if (t) t.focus();
  }
  function closeModal(){
    var m = qs('#connect-modal');
    if (!m) return;
    m.setAttribute('aria-hidden','true');
  }

  document.addEventListener('click', function(e){
    var el = e.target;
    if (!el) return;
    var action = el.getAttribute && el.getAttribute('data-action');
    if (action === 'open-connect-modal') { e.preventDefault(); openModal(); }
    if (action === 'close-connect-modal') { e.preventDefault(); closeModal(); }
  });

  document.addEventListener('keydown', function(e){
    if (e.key === 'Escape') closeModal();
  });
})();
