use tonic::{transport::Server, Request, Response, Status};

pub mod hello_world {
    tonic::include_proto!("urlgen"); // The string specified here must match the proto package name
}
