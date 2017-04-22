var util = {};

util.entityMap = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;',
  '/': '&#x2F;',
  '`': '&#x60;',
  '=': '&#x3D;'
};

util.escapeHtml = function (string) {
  return String(string).replace(/[&<>"'`=\/]/g, function (s) {
    return this.entityMap[s];
  });
};
