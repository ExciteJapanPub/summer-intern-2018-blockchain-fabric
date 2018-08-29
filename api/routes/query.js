'use strict';
const express = require('express');
const { query } = require('express-validator/check');
const validatorModule = require('api-modules/validator');
const fabricQueryModule = require('api-modules/fabric/query');
const router = express.Router();
const csvParser = validatorModule.customSanitizer.parseCSV;

// バリデーション
const validations = [
    query('chaincode', 'chaincode is invalid').exists().isAlphanumeric(),
    query('function', 'fcn is invalid').exists().isAlphanumeric(),
    query('args', 'args is invalid').exists().customSanitizer(csvParser)
];

/**
 * query API
 * @example GET http://localhost:3000/query?chaincode=fabcar&function=queryAllCars&args="[]"
 */
router.get('/', validations, validatorModule.assert, async (req, res) => {
    const fabricModule = new fabricQueryModule(req.query.chaincode);
    const result = await fabricModule.run(req.query.function, req.query.args);

    res.jsonp(result);
});


module.exports = router;
