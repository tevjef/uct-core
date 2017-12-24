package io.coursetrakr.server

import io.grpc.stub.StreamObserver
import com.google.api.client.googleapis.auth.oauth2.GoogleCredential
import com.google.api.client.util.Clock
import io.coursetrakr.proto.TokenRequest
import io.coursetrakr.proto.TokenResponse
import io.coursetrakr.proto.TokenServiceGrpc
import java.io.IOException

class TokenService(private val googleCredential: GoogleCredential)
    : TokenServiceGrpc.TokenServiceImplBase() {

    private var cachedToken = ""

    override fun getToken(request: TokenRequest, responseObserver: StreamObserver<TokenResponse>) {
        try {
            val token = getAccessToken()
            val tokenResponse = TokenResponse.newBuilder()
                    .apply { this.token = token }
                    .build()

            println("sending token...")
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
            println("refreshing token...")
            googleCredential.refreshToken()
        }
        cachedToken = googleCredential.accessToken
        return cachedToken
    }
}