import * as functions from 'firebase-functions';
import {FirebaseAdminService} from "./services/FirebaseAdminService";

// // Start writing Firebase Functions
// // https://firebase.google.com/docs/functions/typescript
//

const firebaseAdminService = new FirebaseAdminService();

export const helloWorld = functions.https.onRequest(async (request, response) => {
    const result = await firebaseAdminService.getLocation().catch((err) => {
        response.json(err);
    });
    response.json(result);
});
