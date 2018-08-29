'use strict';
const express = require('express');
const { body } = require('express-validator/check');
const validatorModule = require('api-modules/validator');
const fabricInvokeModule = require('api-modules/fabric/invoke');
const router = express.Router();
const csvParser = validatorModule.customSanitizer.parseCSV;

// バリデーション
const validations = [
    body('chaincode', 'chaincode is invalid').exists().isAlphanumeric(),
    body('function', 'fcn is invalid').exists().isAlphanumeric(),
    body('args', 'args is invalid').exists().customSanitizer(csvParser)
];

/**
 * invoke API
 * @example POST http://localhost:3000/invoke
 */
router.post('/', validations, validatorModule.assert, async (req, res) => {
    const invoker = new fabricInvokeModule(req.body.chaincode);
    const result = await invoker.run(req.body.function, req.body.args);

    res.jsonp(result);
});


module.exports = router;
