/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CancelablePromise } from '../core/CancelablePromise';
import type { BaseHttpRequest } from '../core/BaseHttpRequest';

export class SystemService {

    constructor(public readonly httpRequest: BaseHttpRequest) {}

    /**
     * Healthcheck
     * Reports version and downstream dependency status. Development only.
     * @returns any OK
     * @throws ApiError
     */
    public getApiHealthcheck(): CancelablePromise<any> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/api/healthcheck',
        });
    }

    /**
     * Checks Readiness
     * Shallow liveness probe. Always returns 200; use /api/healthcheck for dependency status.
     * @returns string OK
     * @throws ApiError
     */
    public getApiReady(): CancelablePromise<Record<string, string>> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/api/ready',
        });
    }

}