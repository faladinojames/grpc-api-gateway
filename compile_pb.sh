protoc \
        --include_imports \
        --include_source_info \
        --descriptor_set_out out.pb \
        pb/authentication.proto pb/gateway.proto pb/payment.proto pb/trip.proto pb/driver-authentication.proto pb/profile.proto

gcloud endpoints services deploy out.pb api_config.yaml