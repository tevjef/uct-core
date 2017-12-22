package io.coursetrakr.server

import io.coursetrackr.proto.TokenRequest
import io.coursetrackr.proto.TokenResponse
import io.coursetrackr.proto.TokenServiceGrpc
import io.grpc.stub.StreamObserver
import com.google.api.client.googleapis.auth.oauth2.GoogleCredential
import com.google.api.client.util.Clock
import java.io.IOException

class TokenService(private val googleCredential: GoogleCredential) : TokenServiceGrpc.TokenServiceImplBase() {
    var cachedToken = ""

    override fun token(request: TokenRequest, responseObserver: StreamObserver<TokenResponse>) {
        try {
            println("Attempting to get token")
            val token = getAccessToken()
            val tokenResponse = TokenResponse.newBuilder()
                    .apply { this.token = token }
                    .build()

            println("Got token")
            responseObserver.onNext(tokenResponse)
            responseObserver.onCompleted()
        } catch (e: Exception) {
            responseObserver.onError(e)
            e.printStackTrace()
        }
    }

    @Throws(IOException::class)
    private fun getAccessToken(): String {
        if (Clock.SYSTEM.currentTimeMillis() > googleCredential.expirationTimeMilliseconds ?: 0) {
            googleCredential.refreshToken()
        }
        cachedToken = googleCredential.accessToken
        return cachedToken
    }
}