# Navigation Engine

Copy this JS block into every deck. Place before `</body>`.

## Standard engine

```javascript
(function() {
  var slides = Array.from(document.querySelectorAll('.slide'));
  var total = slides.length;
  var current = 0;
  var notesVisible = false;

  var progress = document.getElementById('progress');
  var counter = document.getElementById('counter');
  var notesPanel = document.getElementById('notes');
  var notesContent = document.getElementById('notes-content');

  function go(n) {
    if (n < 0 || n >= total) return;
    slides[current].classList.remove('active');
    current = n;
    slides[current].classList.add('active');
    progress.style.width = ((current + 1) / total * 100) + '%';
    counter.textContent = (current + 1) + ' / ' + total;
    notesContent.textContent = slides[current].dataset.notes || '';
  }

  document.addEventListener('keydown', function(e) {
    if (e.key === 'ArrowRight' || e.key === 'ArrowDown' || e.key === ' ') { e.preventDefault(); go(current + 1); }
    else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') { e.preventDefault(); go(current - 1); }
    else if (e.key === 'n' || e.key === 'N') { notesVisible = !notesVisible; notesPanel.classList.toggle('visible', notesVisible); }
    else if (e.key === 'Home') { e.preventDefault(); go(0); }
    else if (e.key === 'End') { e.preventDefault(); go(total - 1); }
  });

  document.addEventListener('click', function(e) {
    if (e.target.closest('.notes-panel')) return;
    if (e.target.closest('a')) return;
    if (e.clientX > window.innerWidth / 2) go(current + 1);
    else go(current - 1);
  });

  go(0);
})();
```

## Required HTML chrome

Place these elements after `.deck` closes but before `<script>`:

```html
<div class="progress-bar" id="progress"></div>
<div class="slide-counter" id="counter"></div>
<div class="key-hint">&larr; &rarr; navigate &middot; N notes</div>
<div class="notes-panel" id="notes">
  <div class="notes-label">Speaker Notes</div>
  <div id="notes-content"></div>
</div>
```

## Slide HTML pattern

```html
<div class="slide" data-notes="Speaker notes go here. Write conversationally — these are what the presenter says, not what the audience reads.">
  <div class="slide-inner">
    <div class="section-label">Section Name</div>
    <div class="t-title mb-md">Slide Title</div>
    <!-- content -->
  </div>
</div>
```

First slide gets `class="slide active"`. All others are just `class="slide"`.

Dark background slides: `class="slide dark-bg"`.

## Controls summary

| Key | Action |
|-----|--------|
| Arrow Right / Down / Space | Next slide |
| Arrow Left / Up | Previous slide |
| N | Toggle speaker notes |
| Home | First slide |
| End | Last slide |
| Click right half | Next slide |
| Click left half | Previous slide |
| Click on `<a>` | Follow link (no navigation) |
