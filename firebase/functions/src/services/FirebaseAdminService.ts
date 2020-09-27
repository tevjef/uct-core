import * as admin from "firebase-admin";
import * as functions from "firebase-functions";

admin.initializeApp();
const auth = admin.auth();
const firestore = admin.firestore();

export class FirebaseAdminService {
  async getUserId(token: string): Promise<string> {
    const value = await auth.verifyIdToken(token, true);
    console.log("TOKEN: " + token);
    console.log("USERID: " + value.uid);

    return value.uid;
  }

  async getLocation(): Promise<any> {
    return firestore
      .collection("main")
      .doc("top")
      .get()
      .then((value) => {
        functions.logger.info("in logs", { data: value.data() });
        if (!value.exists) {
            throw Error("does not exist" + value.ref);
        }
        return (<any>value.data()).field.forEach((val: { get: () => any; }) => {
          return val.get();
        });
      });
  }
}
