#!/usr/bin/env osascript -l JavaScript
// mac-overlay.js — JXA Cocoa overlay notification for macOS
// Usage: osascript -l JavaScript overlay.js <message> <color> <slot> <dismiss_seconds>
//
// Creates a borderless, always-on-top overlay on every screen.
// Dismisses automatically after <dismiss_seconds> seconds.
// No icon. Colors: red, blue, yellow, green, purple, orange.

ObjC.import('Cocoa');

function run(argv) {
  var message = argv[0] || 'notify';
  var color   = argv[1] || 'blue';
  var slot    = parseInt(argv[2], 10) || 0;
  var dismiss = parseFloat(argv[3]) || 4;

  // Color map (r, g, b in 0-255)
  var r = 30, g = 80, b = 180; // default: blue
  switch (color) {
    case 'red':    r = 180; g = 0;   b = 0;   break;
    case 'blue':   r = 30;  g = 80;  b = 180; break;
    case 'yellow': r = 180; g = 140; b = 0;   break;
    case 'green':  r = 20;  g = 140; b = 60;  break;
    case 'purple': r = 100; g = 20;  b = 180; break;
    case 'orange': r = 200; g = 90;  b = 0;   break;
  }

  var bgColor = $.NSColor.colorWithSRGBRedGreenBlueAlpha(r/255, g/255, b/255, 1.0);
  var winWidth = 520, winHeight = 80;

  $.NSApplication.sharedApplication;
  $.NSApp.setActivationPolicy($.NSApplicationActivationPolicyAccessory);

  var screens = $.NSScreen.screens;
  var screenCount = screens.count;

  for (var i = 0; i < screenCount; i++) {
    var screen = screens.objectAtIndex(i);
    var visibleFrame = screen.visibleFrame;

    var yOffset = 40 + slot * 90;
    var x = visibleFrame.origin.x + (visibleFrame.size.width - winWidth) / 2;
    var y = visibleFrame.origin.y + visibleFrame.size.height - winHeight - yOffset;
    var frame = $.NSMakeRect(x, y, winWidth, winHeight);

    var win = $.NSWindow.alloc.initWithContentRectStyleMaskBackingDefer(
      frame,
      $.NSWindowStyleMaskBorderless,
      $.NSBackingStoreBuffered,
      false
    );

    win.setBackgroundColor(bgColor);
    win.setAlphaValue(0.95);
    win.setLevel($.NSStatusWindowLevel);
    win.setIgnoresMouseEvents(true);
    win.setCollectionBehavior(
      $.NSWindowCollectionBehaviorCanJoinAllSpaces |
      $.NSWindowCollectionBehaviorStationary
    );

    win.contentView.wantsLayer = true;
    win.contentView.layer.cornerRadius = 12;
    win.contentView.layer.masksToBounds = true;

    var contentView = win.contentView;
    var font = $.NSFont.boldSystemFontOfSize(18);
    var textHeight = font.ascender - font.descender + font.leading + 4;
    var textY = (winHeight - textHeight) / 2;

    var label = $.NSTextField.alloc.initWithFrame(
      $.NSMakeRect(20, textY, winWidth - 40, textHeight)
    );
    label.setStringValue($(message));
    label.setBezeled(false);
    label.setDrawsBackground(false);
    label.setEditable(false);
    label.setSelectable(false);
    label.setTextColor($.NSColor.whiteColor);
    label.setAlignment($.NSTextAlignmentCenter);
    label.setFont(font);
    label.setLineBreakMode($.NSLineBreakByTruncatingTail);
    label.cell.setWraps(false);
    contentView.addSubview(label);

    win.orderFrontRegardless;
  }

  $.NSTimer.scheduledTimerWithTimeIntervalTargetSelectorUserInfoRepeats(
    dismiss,
    $.NSApp,
    'terminate:',
    null,
    false
  );

  $.NSApp.run;
}
