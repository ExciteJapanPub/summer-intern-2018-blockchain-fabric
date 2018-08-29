'use strict';
const httpStatus = require('http-status-codes');
const { validationResult } = require('express-validator/check');

const Validator = {
    customSanitizer: {
        /**
         * csv形式の文字列を配列にパースするsanitizer
         * @param  {string} value [description]
         * @return {Array}
         */
        parseCSV: (value) => {
            //  未指定なら空の配列を返却
            if (!value) {
                return [];
            }
            return value.split(',');
        }
    },
    /**
     * バリデーション失敗時にエラーレスポンスを返すためのミドルウェア
     * @param  {Object}   req    Express Request Object
     * @param  {Object}   res    Express Response Object
     * @param  {Function} next   次のミドルウェアの呼び出し(引数ありで呼び出すことで、通常のルーティングをスキップしエラーミドルウェアを呼び出す)
     */
    assert: function (req, res, next) {
        try {
            validationResult(req).throw();
            // バリデーションOK
            next();
        } catch (e) {
            const error = {
                status: httpStatus.BAD_REQUEST,
                msg: 'Parameter Error',
                info: e.array()
            };
            // バリデーションNG
            next(error);
        }
    }
};

module.exports = Validator;
