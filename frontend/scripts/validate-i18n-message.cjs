module.exports = function validateMessage(message) {
    const remainder = message.replace(/\{\{[a-zA-Z_][\w-]*\}\}/g, '');

    if (/[{}]/.test(remainder)) {
        throw new SyntaxError('Invalid Gravity i18n placeholder');
    }
};
