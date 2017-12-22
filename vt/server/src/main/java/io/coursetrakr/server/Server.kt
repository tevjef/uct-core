package io.coursetrakr.server

import com.google.api.client.googleapis.auth.oauth2.GoogleCredential
import io.grpc.Server
import io.grpc.ServerBuilder
import java.io.FileInputStream
import java.util.*
import kotlin.system.exitProcess

fun main(args : Array<String>) {
    val scope = "https://www.googleapis.com/auth/firebase.messaging"
    println(System.getProperty("user.dir"))
    val credentialLocation = System.getenv("CREDENTIALS_LOCATION")

    if (credentialLocation.isNullOrBlank()) {
        println("Invalid file location $credentialLocation")
        exitProcess(1)
    }

    val googleCredential = GoogleCredential
            .fromStream(FileInputStream(credentialLocation))
            .createScoped(Arrays.asList(scope))

    val port: Int = System.getenv("VT_PORT")?.toIntOrNull() ?: -1
    if (port == -1) {
        println("VT_PORT environment variable must be provided.")
        exitProcess(1)
    }

    val server: Server = ServerBuilder.forPort(port)
            .addService(TokenService(googleCredential))
            .build()

    println("Starting server on port: $port")
    server.start().awaitTermination()
}


