import { check } from 'k6';
import http from 'k6/http';

export default function () {
    const url = 'http://localhost:4466';
    //const payload = `{"query":"mutation CreatePost {createOnePost(data: {title: \\"myPost\\" author: {connect: {email: \\"jens@wundergraph.com\\"}}}){id title}}","operationName":"CreatePost"}`;
    const payload = `{"query":"query AllPosts {findManyPost(take: 2500){id title createdAt}}","operationName":"CreatePost"}`;

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(url, payload, params);
    check(res, {
        'is status 200': (r) => r.status === 200,
    });
    check(res, {
        'verify body': (r) =>
            r.body.includes('myPost'),
    });
}
