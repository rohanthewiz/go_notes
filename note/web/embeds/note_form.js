
require.config({ paths: { 'vs': 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.2/min/vs' }});
require(['vs/editor/editor.main'], function() {

    // Make our own theme
    monaco.editor.defineTheme('ro-dark', {
        base: 'vs-dark',
        inherit: true,
        rules: [
            { background: '1d1f21' },
            { token: 'comment', foreground: '909090' },
            { token: 'string', foreground: 'b5bd68' },
            { token: 'variable', foreground: 'c5c8c6' },
            { token: 'keyword', foreground: 'ba7d57' },
            { token: 'number', foreground: 'de935f' },
        ],
        colors: {
            'editorBackground': '#1d1f21',
            // 'editorForeground': '#c5c8c6',
            // 'editor.selectionBackground': '#373b41',
            'editorCursor.foreground': '#6DDADA',
            'editor.lineHighlightBackground': '#606060',
        }
    });

    // var init_val = document.getElementById("note_body").value;
    var editor = monaco.editor.create(document.getElementById('editor'), {
        value: window.codeObj.Code,
        language: 'markdown',
        theme: 'ro-dark',
        lineNumbers: 'on',
        minimap: {
            enabled: false
        },
        renderLineHighlight: 'gutter'
    });

    var nf = document.getElementById("note_form");

    nf.addEventListener("submit", function() {
        var nb = document.getElementById("note_body");
        if (typeof editor !== 'undefined' && nb !== null && typeof nb !== 'undefined') {
            nb.value = editor.getValue();
        }
        return true;
    });

    // Optionally add a custom keyboard shortcut
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyP, function() {
        editor.trigger('keyboard', 'editor.action.quickCommand');
    });
});
